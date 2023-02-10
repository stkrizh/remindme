package deps

import (
	"context"
	"remindme/internal/config"
	"remindme/internal/core/domain/channel"
	c "remindme/internal/core/domain/common"
	dl "remindme/internal/core/domain/logging"
	drl "remindme/internal/core/domain/rate_limiter"
	"remindme/internal/core/domain/reminder"
	duow "remindme/internal/core/domain/unit_of_work"
	"remindme/internal/core/domain/user"
	sendreminder "remindme/internal/core/services/send_reminder"
	dbchannel "remindme/internal/db/channel"
	dbreminder "remindme/internal/db/reminder"
	uow "remindme/internal/db/unit_of_work"
	dbuser "remindme/internal/db/user"
	"remindme/internal/implementations/logging"
	passwordhasher "remindme/internal/implementations/password_hasher"
	passwordresetter "remindme/internal/implementations/password_resetter"
	randomstringgenerator "remindme/internal/implementations/random_string_generator"
	ratelimiter "remindme/internal/implementations/rate_limiter"
	remindersender "remindme/internal/implementations/reminder_sender"
	telegrambotmessagesender "remindme/internal/implementations/telegram_bot_message_sender"
	"remindme/internal/rabbitmq"
	reminderreadyforsending "remindme/internal/rabbitmq/consumers/reminder_ready_for_sending"
	reminderscheduler "remindme/internal/rabbitmq/publishers/reminder_scheduler"
	"sync"
	"time"

	"github.com/go-redis/redis/v9"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/rabbitmq/amqp091-go"
)

type Deps struct {
	Config *config.Config
	Logger dl.Logger

	DB       *pgxpool.Pool
	Redis    *redis.Client
	Rabbitmq *rabbitmq.Connection

	Now func() time.Time

	UnitOfWork         duow.UnitOfWork
	UserRepository     user.UserRepository
	SessionRepository  user.SessionRepository
	ChannelRepository  channel.Repository
	ReminderRepository reminder.ReminderRepository

	RateLimiter drl.RateLimiter

	UserActivationTokenGenerator user.ActivationTokenGenerator
	UserActivationTokenSender    user.ActivationTokenSender
	UserIdentityGenerator        user.IdentityGenerator
	UserSessionTokenGenerator    user.SessionTokenGenerator
	PasswordHasher               user.PasswordHasher
	PasswordResetter             user.PasswordResetter
	PasswordResetTokenSender     user.PasswordResetTokenSender
	DefaultUserLimits            user.Limits
	DefaultAnonymousUserLimits   user.Limits

	ChannelVerificationTokenGenerator channel.VerificationTokenGenerator

	ReminderScheduler reminder.Scheduler
	ReminderSender    reminder.Sender

	TelegramBotMessageSender *telegrambotmessagesender.TelegramBotMessageSender
}

func InitDeps() (*Deps, func()) {
	deps := &Deps{}

	deps.initConfig()
	closeLogger := deps.initLogger()
	closePgxPool := deps.initPgxPool()
	closeRedisClient := deps.initRedisClient()
	closeRabbitmqConn := deps.initRabbitmqConnection()

	deps.UnitOfWork = uow.NewPgxUnitOfWork(deps.DB)
	deps.UserRepository = dbuser.NewPgxRepository(deps.DB)
	deps.SessionRepository = dbuser.NewPgxSessionRepository(deps.DB)
	deps.ChannelRepository = dbchannel.NewPgxChannelRepository(deps.DB)
	deps.ReminderRepository = dbreminder.NewPgxReminderRepository(deps.DB)

	deps.Now = func() time.Time { return time.Now().UTC() }
	deps.RateLimiter = ratelimiter.NewRedis(deps.Redis, deps.Logger, deps.Now)
	deps.UserActivationTokenGenerator = randomstringgenerator.NewGenerator()
	deps.UserActivationTokenSender = user.NewFakeActivationTokenSender()
	deps.UserIdentityGenerator = randomstringgenerator.NewGenerator()
	deps.UserSessionTokenGenerator = randomstringgenerator.NewGenerator()
	deps.PasswordHasher = passwordhasher.NewBcrypt(deps.Config.Secret, deps.Config.BcryptHasherCost)
	deps.PasswordResetter = passwordresetter.NewHMAC(
		deps.Config.Secret,
		time.Duration(deps.Config.PasswordResetValidDurationHours*int(time.Hour)),
		deps.Now,
	)
	deps.PasswordResetTokenSender = user.NewFakePasswordResetTokenSender()
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
	deps.ReminderSender = remindersender.New(deps.Logger)

	deps.TelegramBotMessageSender = telegrambotmessagesender.New(
		deps.Config.TelegramBaseURL,
		deps.Config.TelegramTokenByBot(),
		deps.Config.TelegramRequestTimeout,
	)

	closeReminderReadyForSendingConsumer := deps.initReminderReadyForSendingConsumer()

	return deps, func() {
		closeFuncs := []func(){
			closeLogger,
			closePgxPool,
			closeRedisClient,
			closeRabbitmqConn,
			closeReminderScheduler,
			closeReminderReadyForSendingConsumer,
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
	return func() { db.Close() }
}

func (deps *Deps) initRedisClient() func() {
	redisOpt, err := redis.ParseURL(deps.Config.RedisURL)
	if err != nil {
		deps.Logger.Error(context.Background(), "Could not connect to Redis.", dl.Entry("err", err))
		panic(err)
	}
	redisClient := redis.NewClient(redisOpt)
	deps.Redis = redisClient
	return func() { redisClient.Close() }
}

func (deps *Deps) initRabbitmqConnection() func() {
	rabbitmqConnection, err := rabbitmq.Dial(deps.Config.RabbitmqURL, deps.Logger)
	if err != nil {
		deps.Logger.Error(context.Background(), "Could not connect to RabbitMQ.", dl.Entry("err", err))
		panic("could not connect to RabbitMQ")
	}
	deps.Rabbitmq = rabbitmqConnection
	return func() { rabbitmqConnection.Close() }
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

	return func() { rabbitmqChannel.Close() }
}

func (deps *Deps) initReminderReadyForSendingConsumer() func() {
	rabbitmqChannel, err := deps.Rabbitmq.Channel()
	if err != nil {
		deps.Logger.Error(context.Background(), "Could not create RabbitMQ channel.", dl.Entry("err", err))
		panic(err)
	}
	reminderReadyForSendingConsumer := reminderreadyforsending.New(
		deps.Logger,
		rabbitmqChannel,
		deps.Config.RabbitmqReminderReadyQueue,
		sendreminder.NewSendService(
			deps.Logger,
			deps.ReminderRepository,
			deps.ReminderSender,
			deps.Now,
			sendreminder.NewPrepareService(
				deps.Logger,
				deps.UnitOfWork,
				deps.Now,
			),
		),
	)
	if err = reminderReadyForSendingConsumer.Consume(); err != nil {
		deps.Logger.Error(context.Background(), "Could not start RabbitMQ consuming.", dl.Entry("err", err))
		panic(err)
	}
	return func() { rabbitmqChannel.Close() }
}
