package deps

import (
	"context"
	"fmt"
	"remindme/internal/config"
	"remindme/internal/core/domain/bot"
	"remindme/internal/core/domain/channel"
	c "remindme/internal/core/domain/common"
	dl "remindme/internal/core/domain/logging"
	drl "remindme/internal/core/domain/rate_limiter"
	"remindme/internal/core/domain/reminder"
	duow "remindme/internal/core/domain/unit_of_work"
	"remindme/internal/core/domain/user"
	"remindme/internal/core/services/captcha"
	dbchannel "remindme/internal/db/channel"
	dbreminder "remindme/internal/db/reminder"
	uow "remindme/internal/db/unit_of_work"
	dbuser "remindme/internal/db/user"
	"remindme/internal/implementations/email"
	"remindme/internal/implementations/logging"
	passwordhasher "remindme/internal/implementations/password_hasher"
	passwordresetter "remindme/internal/implementations/password_resetter"
	randomstringgenerator "remindme/internal/implementations/random_string_generator"
	ratelimiter "remindme/internal/implementations/rate_limiter"
	recaptcha "remindme/internal/implementations/recaptcha"
	remindernlqparser "remindme/internal/implementations/reminder_nlq_parser"
	remindersender "remindme/internal/implementations/reminder_sender"
	telegrambotmessagesender "remindme/internal/implementations/telegram_bot_message_sender"
	"remindme/internal/rabbitmq"
	reminderscheduler "remindme/internal/rabbitmq/publishers/reminder_scheduler"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/retry"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/getsentry/sentry-go"
	"github.com/go-redis/redis/v9"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/r3labs/sse/v2"
	"github.com/rabbitmq/amqp091-go"
)

type Deps struct {
	Config    *config.Config
	AwsConfig aws.Config
	Logger    dl.Logger

	DB        *pgxpool.Pool
	Redis     *redis.Client
	Rabbitmq  *rabbitmq.Connection
	SseServer *sse.Server

	Now func() time.Time

	UnitOfWork         duow.UnitOfWork
	UserRepository     user.UserRepository
	LimitsRepository   user.LimitsRepository
	SessionRepository  user.SessionRepository
	ChannelRepository  channel.Repository
	ReminderRepository reminder.ReminderRepository

	RateLimiter drl.RateLimiter

	EmailSender              *email.EmailSender
	TelegramBotMessageSender bot.TelegramBotMessageSender

	UserActivationTokenGenerator user.ActivationTokenGenerator
	UserActivationTokenSender    user.ActivationTokenSender
	UserIdentityGenerator        user.IdentityGenerator
	UserSessionTokenGenerator    user.SessionTokenGenerator
	PasswordHasher               user.PasswordHasher
	PasswordResetter             user.PasswordResetter
	PasswordResetTokenSender     user.PasswordResetTokenSender
	CaptchaValidator             captcha.CaptchaValidator
	DefaultUserLimits            user.Limits
	DefaultAnonymousUserLimits   user.Limits

	ChannelVerificationTokenGenerator channel.VerificationTokenGenerator

	ReminderScheduler reminder.Scheduler
	ReminderSender    reminder.Sender
	ReminderNLQParser reminder.NaturalLanguageQueryParser
}

func InitDeps() (*Deps, func()) {
	deps := &Deps{}

	deps.initConfig()
	deps.initAwsConfig()

	closeLogger := deps.initLogger()
	closePgxPool := deps.initPgxPool()
	closeRedisClient := deps.initRedisClient()
	closeRabbitmqConn := deps.initRabbitmqConnection()
	closeSseServer := deps.initSseServer()

	deps.UnitOfWork = uow.NewPgxUnitOfWork(deps.DB)
	deps.UserRepository = dbuser.NewPgxRepository(deps.DB)
	deps.LimitsRepository = dbuser.NewPgxLimitsRepository(deps.DB)
	deps.SessionRepository = dbuser.NewPgxSessionRepository(deps.DB)
	deps.ChannelRepository = dbchannel.NewPgxChannelRepository(deps.DB)
	deps.ReminderRepository = dbreminder.NewPgxReminderRepository(deps.DB)

	deps.EmailSender = email.NewEmailSender(
		deps.AwsConfig,
		deps.Config.AwsEmailSender,
		deps.Config.AwsEmailActivateAccountTemplate,
		deps.Config.AwsEmailActivationUrl,
		deps.Config.AwsEmailPasswordResetTemplate,
		deps.Config.AwsEmailPasswordResetBaseUrl,
		deps.Config.AwsEmailActivateChannelTemplate,
	)

	deps.Now = func() time.Time { return time.Now().UTC() }
	deps.RateLimiter = ratelimiter.NewRedis(deps.Redis, deps.Logger, deps.Now)
	deps.UserActivationTokenGenerator = randomstringgenerator.NewGenerator()
	deps.UserActivationTokenSender = deps.EmailSender
	deps.UserIdentityGenerator = randomstringgenerator.NewGenerator()
	deps.UserSessionTokenGenerator = randomstringgenerator.NewGenerator()
	deps.PasswordHasher = passwordhasher.NewBcrypt(deps.Config.Secret, deps.Config.BcryptHasherCost)
	deps.PasswordResetter = passwordresetter.NewHMAC(
		deps.Config.Secret,
		time.Duration(deps.Config.PasswordResetValidDurationHours*int(time.Hour)),
		deps.Now,
	)
	deps.PasswordResetTokenSender = deps.EmailSender
	deps.CaptchaValidator = deps.initCaptchaValidator()
	deps.DefaultUserLimits = user.Limits{
		EmailChannelCount:        c.NewOptional(uint32(1), true),
		TelegramChannelCount:     c.NewOptional(uint32(1), true),
		ActiveReminderCount:      c.NewOptional(uint32(10), true),
		MonthlySentReminderCount: c.NewOptional(uint32(100), true),
		ReminderEveryPerDayCount: c.NewOptional(1.0, true),
	}
	deps.DefaultAnonymousUserLimits = user.Limits{
		EmailChannelCount:        c.NewOptional(uint32(1), true),
		TelegramChannelCount:     c.NewOptional(uint32(1), true),
		ActiveReminderCount:      c.NewOptional(uint32(5), true),
		MonthlySentReminderCount: c.NewOptional(uint32(50), true),
		ReminderEveryPerDayCount: c.NewOptional(1.0, true),
	}

	deps.ChannelVerificationTokenGenerator = randomstringgenerator.NewGenerator()

	closeReminderScheduler := deps.initRabbitmqReminderScheduler()

	deps.TelegramBotMessageSender = telegrambotmessagesender.New(
		deps.Config.TelegramBaseURL,
		deps.Config.TelegramTokenByBot(),
		deps.Config.TelegramRequestTimeout,
	)

	deps.ReminderSender = remindersender.New(
		deps.Logger,
		deps.ChannelRepository,
		deps.SseServer,
		remindersender.NewEmail(deps.AwsConfig, deps.Config.AwsEmailSender, deps.Config.AwsEmailReminderTemplate),
		remindersender.NewTelegram(deps.TelegramBotMessageSender),
		remindersender.NewInternal(deps.SseServer),
	)
	deps.ReminderNLQParser = remindernlqparser.New()

	flushSentry := deps.initSentry()

	return deps, func() {
		closeFuncs := []func(){
			closeSseServer,
			closeReminderScheduler,
			closeRabbitmqConn,
			closeRedisClient,
			closePgxPool,
			closeLogger,
			flushSentry,
		}

		var wg sync.WaitGroup
		wg.Add(len(closeFuncs))
		for _, closeFunc := range closeFuncs {
			closeFunc := closeFunc
			go func() {
				closeFunc()
				wg.Done()
			}()
		}

		wg.Wait()
	}
}

func (deps *Deps) initConfig() {
	config, err := config.Load()
	if err != nil {
		panic(err)
	}
	deps.Config = config
}

func (deps *Deps) initAwsConfig() {
	cfg, err := awsConfig.LoadDefaultConfig(
		context.Background(),
		awsConfig.WithRegion(deps.Config.AwsRegion),
		awsConfig.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(
				deps.Config.AwsAccessKey,
				deps.Config.AwsSecretKey,
				"",
			),
		),
		awsConfig.WithRetryer(func() aws.Retryer {
			return retry.AddWithMaxAttempts(
				retry.AddWithMaxBackoffDelay(retry.NewStandard(), time.Second*5),
				3,
			)
		}),
	)
	if err != nil {
		panic(err)
	}
	deps.AwsConfig = cfg
}

func (deps *Deps) initLogger() func() {
	logger := logging.NewZapLogger()
	deps.Logger = logger
	return func() { logger.Sync() }
}

func (deps *Deps) initPgxPool() func() {
	db, err := pgxpool.Connect(context.Background(), deps.Config.PostgresqlURL)
	if err != nil {
		deps.Logger.Error(context.Background(), "Could not connect to DB.", dl.Entry("err", err))
		panic(err)
	}
	deps.DB = db
	return func() {
		deps.Logger.Info(context.Background(), "Shutting down DB connection.")
		db.Close()
		deps.Logger.Info(context.Background(), "DB connection shut down.")
	}
}

func (deps *Deps) initRedisClient() func() {
	redisOpt, err := redis.ParseURL(deps.Config.RedisURL)
	if err != nil {
		deps.Logger.Error(context.Background(), "Could not connect to Redis.", dl.Entry("err", err))
		panic(err)
	}
	redisClient := redis.NewClient(redisOpt)
	deps.Redis = redisClient
	return func() {
		deps.Logger.Info(context.Background(), "Shutting down Redis client.")
		redisClient.Close()
		deps.Logger.Info(context.Background(), "Redis client shut down.")
	}
}

func (deps *Deps) initRabbitmqConnection() func() {
	rabbitmqConnection, err := rabbitmq.Dial(deps.Config.RabbitmqURL, deps.Logger)
	if err != nil {
		deps.Logger.Error(context.Background(), "Could not connect to RabbitMQ.", dl.Entry("err", err))
		panic("could not connect to RabbitMQ")
	}
	deps.Rabbitmq = rabbitmqConnection
	return func() {
		deps.Logger.Info(context.Background(), "Shutting down RabbitMQ connection.")
		rabbitmqConnection.Close()
		deps.Logger.Info(context.Background(), "RabbitMQ connection shut down.")
	}
}

func (deps *Deps) initRabbitmqReminderScheduler() func() {
	rabbitmqChannel, err := deps.Rabbitmq.Channel()
	if err != nil {
		deps.Logger.Error(context.Background(), "Could not create RabbitMQ channel.", dl.Entry("err", err))
		panic(err)
	}

	err = rabbitmqChannel.ExchangeDeclare(
		deps.Config.RabbitmqDelayedExchange,
		"x-delayed-message",
		true,
		false,
		false,
		false,
		amqp091.Table{"x-delayed-type": "direct"},
	)
	if err != nil {
		deps.Logger.Error(context.Background(), "Could not create RabbitMQ exhange.", dl.Entry("err", err))
		panic(err)
	}
	_, err = rabbitmqChannel.QueueDeclare(deps.Config.RabbitmqReminderReadyQueue, true, false, false, false, nil)
	if err != nil {
		deps.Logger.Error(context.Background(), "Could not create RabbitMQ queue.", dl.Entry("err", err))
		panic(err)
	}
	if err := rabbitmqChannel.QueueBind(
		deps.Config.RabbitmqReminderReadyQueue,
		deps.Config.RabbitmqReminderReadyQueue,
		deps.Config.RabbitmqDelayedExchange,
		false,
		nil,
	); err != nil {
		deps.Logger.Error(context.Background(), "Could not bind queue to RabbitMQ exhange.", dl.Entry("err", err))
		panic(err)
	}

	deps.ReminderScheduler = reminderscheduler.NewRabbitMQ(
		deps.Logger,
		rabbitmqChannel,
		deps.Config.RabbitmqDelayedExchange,
		deps.Config.RabbitmqReminderReadyQueue,
		deps.Now,
	)

	return func() {
		deps.Logger.Info(context.Background(), "Shutting down reminder scheduller.")
		rabbitmqChannel.Close()
		deps.Logger.Info(context.Background(), "Reminder scheduller shut down.")
	}
}

func (deps *Deps) initSseServer() func() {
	deps.SseServer = sse.New()
	deps.SseServer.AutoStream = true
	deps.SseServer.AutoReplay = false
	return func() {
		deps.Logger.Info(context.Background(), "Shutting down SSE server.")
		deps.SseServer.Close()
		deps.Logger.Info(context.Background(), "SSE server shut down.")
	}
}

func (deps *Deps) initCaptchaValidator() captcha.CaptchaValidator {
	if deps.Config.IsTestMode {
		return captcha.NewAllowAlwaysCaptchaValidator()
	}
	return recaptcha.New(
		deps.Logger,
		deps.Config.GoogleRecaptchaSecretKey,
		deps.Config.GoogleRecaptchaScoreThreshold,
		deps.Config.GoogleRecaptchaRequestTimeout,
	)
}

func (deps *Deps) initSentry() func() {
	if deps.Config.SentryDsn != nil {
		err := sentry.Init(sentry.ClientOptions{
			Dsn:              deps.Config.SentryDsn.String(),
			TracesSampleRate: 0.01,
		})
		if err != nil {
			panic(fmt.Sprintf("could not init Sentry: %v\n", err))
		}
		deps.Logger.Info(context.Background(), "Sentry has been successfully initialized.")
		return func() {
			ok := sentry.Flush(5 * time.Second)
			deps.Logger.Info(context.Background(), "Sentry events flushed.", dl.Entry("ok", ok))
		}
	}

	deps.Logger.Info(context.Background(), "Sentry is disabled.")
	return func() {}
}
