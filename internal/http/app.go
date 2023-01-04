package http

import (
	"context"
	"fmt"
	"net/http"
	"remindme/internal/config"
	"remindme/internal/core/domain/channel"
	c "remindme/internal/core/domain/common"
	dl "remindme/internal/core/domain/logging"
	drl "remindme/internal/core/domain/rate_limiter"
	"remindme/internal/core/domain/user"
	serviceActivateUser "remindme/internal/core/services/activate_user"
	serviceAuth "remindme/internal/core/services/auth"
	serviceCreateEmailChannel "remindme/internal/core/services/create_email_channel"
	serviceCreateTelegramChannel "remindme/internal/core/services/create_telegram_channel"
	serviceGetUserBySessionToken "remindme/internal/core/services/get_user_by_session_token"
	serviceListUserChannels "remindme/internal/core/services/list_user_channels"
	serviceLogInWithEmail "remindme/internal/core/services/log_in_with_email"
	serviceLogOut "remindme/internal/core/services/log_out"
	serviceRateLimiting "remindme/internal/core/services/rate_limiting"
	serviceResetPassword "remindme/internal/core/services/reset_password"
	serviceSendPasswordResetToken "remindme/internal/core/services/send_password_reset_token"
	serviceSignUpAnonymously "remindme/internal/core/services/sign_up_anonymously"
	serviceSignUpWithEmail "remindme/internal/core/services/sign_up_with_email"
	serviceVerifyEmailChannel "remindme/internal/core/services/verify_email_channel"
	serviceVerifyTelegramChannel "remindme/internal/core/services/verify_telegram_channel"
	dbchannel "remindme/internal/db/channel"
	uow "remindme/internal/db/unit_of_work"
	dbuser "remindme/internal/db/user"
	authMiddleware "remindme/internal/http/handlers/auth"
	handlerActivate "remindme/internal/http/handlers/auth/activate_user"
	handlerLogin "remindme/internal/http/handlers/auth/log_in_with_email"
	handlerLogout "remindme/internal/http/handlers/auth/log_out"
	handlerMe "remindme/internal/http/handlers/auth/me"
	handlerResetPassword "remindme/internal/http/handlers/auth/reset_password"
	handlerSendPasswResetToken "remindme/internal/http/handlers/auth/send_password_reset_token"
	handlerSignUpAnon "remindme/internal/http/handlers/auth/sign_up_anonymously"
	handlerSignUpWithEmail "remindme/internal/http/handlers/auth/sign_up_with_email"
	handlerCreateEmailChannel "remindme/internal/http/handlers/channels/create_email_channel"
	handlerCreateTelegramChannel "remindme/internal/http/handlers/channels/create_telegram_channel"
	handlerListUserChannels "remindme/internal/http/handlers/channels/list_user_channels"
	handlerVerifyEmailChannel "remindme/internal/http/handlers/channels/verify_email_channel"
	handlerTelegramUpdates "remindme/internal/http/handlers/telegram"
	"remindme/internal/implementations/logging"
	passwordhasher "remindme/internal/implementations/password_hasher"
	passwordresetter "remindme/internal/implementations/password_resetter"
	randomstringgenerator "remindme/internal/implementations/random_string_generator"
	ratelimiter "remindme/internal/implementations/rate_limiter"
	telegrambotmessagesender "remindme/internal/implementations/telegram_bot_message_sender"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-redis/redis/v9"
	"github.com/jackc/pgx/v4/pgxpool"
)

func StartApp() {
	config, err := config.Load()
	if err != nil {
		panic(err)
	}

	db, err := pgxpool.Connect(context.Background(), config.PostgresqlURL)
	if err != nil {
		panic("Could not connect to the database.")
	}
	defer db.Close()

	redisOpt, err := redis.ParseURL(config.RedisURL)
	if err != nil {
		panic(err)
	}
	redisClient := redis.NewClient(redisOpt)

	now := func() time.Time { return time.Now().UTC() }

	logger := logging.NewZapLogger()
	defer logger.Sync()

	unitOfWork := uow.NewPgxUnitOfWork(db)
	userRepository := dbuser.NewPgxRepository(db)
	sessionRepository := dbuser.NewPgxSessionRepository(db)
	channelRepository := dbchannel.NewPgxChannelRepository(db)
	rateLimiter := ratelimiter.NewRedis(redisClient, logger, now)
	randomStringGenerator := randomstringgenerator.NewGenerator()

	passwordHasher := passwordhasher.NewBcrypt(config.Secret, config.BcryptHasherCost)
	activationTokenSender := user.NewFakeActivationTokenSender()
	passwordResetter := passwordresetter.NewHMAC(
		config.Secret,
		time.Duration(config.PasswordResetValidDurationHours*int(time.Hour)),
		now,
	)
	passwordResetTokenSender := user.NewFakePasswordResetTokenSender()
	defaultUserLimits := user.Limits{
		EmailChannelCount:        c.NewOptional(uint32(1), true),
		TelegramChannelCount:     c.NewOptional(uint32(1), true),
		ActiveReminderCount:      c.NewOptional(uint32(10), true),
		MonthlySentReminderCount: c.NewOptional(uint32(100), true),
		ReminderEveryPerDayCount: c.NewOptional(1.0, true),
	}
	defaultAnonymousUserLimits := user.Limits{
		EmailChannelCount:        c.NewOptional(uint32(1), true),
		TelegramChannelCount:     c.NewOptional(uint32(1), true),
		ActiveReminderCount:      c.NewOptional(uint32(5), true),
		MonthlySentReminderCount: c.NewOptional(uint32(50), true),
		ReminderEveryPerDayCount: c.NewOptional(1.0, true),
	}

	telegramBotSender := telegrambotmessagesender.New(
		config.TelegramBaseURL,
		config.TelegramTokenByBot(),
		config.TelegramRequestTimeout,
	)

	signUpWithEmailService := serviceSignUpWithEmail.NewWithActivationTokenSending(
		logger,
		activationTokenSender,
		serviceSignUpWithEmail.New(
			logger,
			unitOfWork,
			passwordHasher,
			randomStringGenerator,
			now,
		),
	)
	signUpAnonymouslyService := serviceSignUpAnonymously.New(
		logger,
		unitOfWork,
		randomStringGenerator,
		randomStringGenerator,
		now,
		defaultAnonymousUserLimits,
	)
	activateUserService := serviceActivateUser.New(
		logger,
		unitOfWork,
		now,
		defaultUserLimits,
	)
	logInWithEmailService := serviceRateLimiting.WithRateLimiting(
		logger,
		rateLimiter,
		drl.Limit{Interval: drl.Hour, Value: 10},
		serviceLogInWithEmail.New(
			logger,
			userRepository,
			sessionRepository,
			passwordHasher,
			randomStringGenerator,
			now,
		),
	)
	logOutService := serviceLogOut.New(
		logger,
		sessionRepository,
	)
	sendPasswordResetToken := serviceRateLimiting.WithRateLimiting(
		logger,
		rateLimiter,
		drl.Limit{Interval: drl.Hour, Value: 3},
		serviceSendPasswordResetToken.New(
			logger,
			userRepository,
			passwordResetter,
			passwordResetTokenSender,
		),
	)
	resetPassword := serviceResetPassword.New(
		logger,
		userRepository,
		passwordResetter,
		passwordHasher,
	)

	getUserBySessionTokenService := serviceAuth.WithAuthentication(
		sessionRepository,
		serviceGetUserBySessionToken.New(),
	)

	createEmailChannel := serviceAuth.WithAuthentication(
		sessionRepository,
		serviceCreateEmailChannel.NewWithVerificationTokenSending(
			logger,
			channel.NewFakeVerificationTokenSender(),
			serviceCreateEmailChannel.New(
				logger,
				unitOfWork,
				randomStringGenerator,
				now,
			),
		),
	)
	verifyEmailChannel := serviceAuth.WithAuthentication(
		sessionRepository,
		serviceRateLimiting.WithRateLimiting(
			logger,
			rateLimiter,
			drl.Limit{Interval: drl.Minute, Value: 5},
			serviceVerifyEmailChannel.New(
				logger,
				channelRepository,
				now,
			),
		),
	)
	createTelegramChannel := serviceAuth.WithAuthentication(
		sessionRepository,
		serviceCreateTelegramChannel.New(
			logger,
			unitOfWork,
			randomStringGenerator,
			now,
		),
	)
	verifyTelegramChannel := serviceVerifyTelegramChannel.New(
		logger,
		channelRepository,
		now,
	)
	listUserChannels := serviceAuth.WithAuthentication(
		sessionRepository,
		serviceListUserChannels.New(
			logger,
			channelRepository,
		),
	)

	authRouter := chi.NewRouter()
	authRouter.Method(http.MethodPost, "/signup", handlerSignUpWithEmail.New(signUpWithEmailService, config.IsTestMode))
	authRouter.Method(http.MethodPost, "/signup/anonymously", handlerSignUpAnon.New(signUpAnonymouslyService))
	authRouter.Method(http.MethodPost, "/activate", handlerActivate.New(activateUserService))
	authRouter.Method(http.MethodPost, "/login", handlerLogin.New(logInWithEmailService))
	authRouter.Method(http.MethodPost, "/logout", handlerLogout.New(logOutService))
	authRouter.Method(
		http.MethodPost,
		"/password_reset/send",
		handlerSendPasswResetToken.New(sendPasswordResetToken, config.IsTestMode),
	)
	authRouter.Method(http.MethodPost, "/password_reset", handlerResetPassword.New(resetPassword))

	profileRouter := chi.NewRouter()
	profileRouter.Use(authMiddleware.SetAuthTokenToContext)
	profileRouter.Method(http.MethodGet, "/me", handlerMe.New(getUserBySessionTokenService))

	channelsRouter := chi.NewRouter()
	channelsRouter.Use(authMiddleware.SetAuthTokenToContext)
	channelsRouter.Method(
		http.MethodGet,
		"/",
		handlerListUserChannels.New(listUserChannels),
	)
	channelsRouter.Method(
		http.MethodPost,
		"/email",
		handlerCreateEmailChannel.New(createEmailChannel, config.IsTestMode),
	)
	channelsRouter.Method(
		http.MethodPost,
		"/telegram",
		handlerCreateTelegramChannel.New(createTelegramChannel, channel.TelegramBot(config.TelegramBots[0])),
	)
	channelsRouter.Method(
		http.MethodPut,
		"/{channelID:[0-9]+}/verification",
		handlerVerifyEmailChannel.New(verifyEmailChannel),
	)

	telegramRouter := chi.NewRouter()
	telegramRouter.Method(
		http.MethodPost,
		fmt.Sprintf("/updates/{bot}/%s", config.TelegramURLSecret),
		handlerTelegramUpdates.New(logger, telegramBotSender, verifyTelegramChannel),
	)

	router := chi.NewRouter()
	router.Mount("/auth", authRouter)
	router.Mount("/profile", profileRouter)
	router.Mount("/channels", channelsRouter)
	router.Mount("/telegram", telegramRouter)

	address := fmt.Sprintf("0.0.0.0:%d", config.Port)

	srv := &http.Server{
		Handler:           router,
		Addr:              address,
		ReadTimeout:       5 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      5 * time.Second,
		IdleTimeout:       5 * time.Second,
	}

	logger.Info(
		context.Background(),
		"Server has started.",
		dl.Entry("address", address),
		dl.Entry("isTestMode", config.IsTestMode),
		dl.Entry("telegramBots", config.TelegramBots),
	)
	srv.ListenAndServe()
}
