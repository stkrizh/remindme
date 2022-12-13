package http

import (
	"context"
	"net/http"
	"remindme/internal/config"
	dl "remindme/internal/core/domain/logging"
	drl "remindme/internal/core/domain/rate_limiter"
	"remindme/internal/core/domain/user"
	activateuser "remindme/internal/core/services/activate_user"
	loginwithemail "remindme/internal/core/services/log_in_with_email"
	logout "remindme/internal/core/services/log_out"
	ratelimiting "remindme/internal/core/services/rate_limiting"
	resetpassword "remindme/internal/core/services/reset_password"
	sendpasswordresettoken "remindme/internal/core/services/send_password_reset_token"
	signupanonymously "remindme/internal/core/services/sign_up_anonymously"
	signupwithemail "remindme/internal/core/services/sign_up_with_email"
	uow "remindme/internal/db/unit_of_work"
	dbuser "remindme/internal/db/user"
	"remindme/internal/http/handlers"
	"remindme/internal/implementations/activation"
	"remindme/internal/implementations/identity"
	"remindme/internal/implementations/logging"
	passwordhasher "remindme/internal/implementations/password_hasher"
	passwordresetter "remindme/internal/implementations/password_resetter"
	ratelimiter "remindme/internal/implementations/rate_limiter"
	"remindme/internal/implementations/session"
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
	rateLimiter := ratelimiter.NewRedis(redisClient, logger, now)

	passwordHasher := passwordhasher.NewBcrypt(config.Secret, config.BcryptHasherCost)
	activationTokenGenerator := activation.NewTokenGenerator()
	activationTokenSender := user.NewFakeActivationTokenSender()
	identityGenerator := identity.NewUUID()
	sessionTokenGenerator := session.NewUUID()
	passwordResetter := passwordresetter.NewHMAC(
		config.Secret,
		time.Duration(config.PasswordResetValidDurationHours*int(time.Hour)),
		now,
	)
	passwordResetTokenSender := user.NewFakePasswordResetTokenSender()

	signUpWithEmailService := signupwithemail.NewWithActivationTokenSending(
		logger,
		activationTokenSender,
		signupwithemail.New(
			logger,
			unitOfWork,
			passwordHasher,
			activationTokenGenerator,
			now,
		),
	)
	signUpAnonymouslyService := signupanonymously.New(
		logger,
		unitOfWork,
		identityGenerator,
		sessionTokenGenerator,
		now,
	)
	activateUserService := activateuser.New(
		logger,
		userRepository,
		now,
	)
	logInWithEmailService := ratelimiting.New(
		logger,
		rateLimiter,
		drl.Limit{Interval: drl.Hour, Value: 10},
		loginwithemail.New(
			logger,
			userRepository,
			sessionRepository,
			passwordHasher,
			sessionTokenGenerator,
			now,
		),
	)
	logOutService := logout.New(
		logger,
		sessionRepository,
	)
	sendPasswordResetToken := ratelimiting.New(
		logger,
		rateLimiter,
		drl.Limit{Interval: drl.Hour, Value: 3},
		sendpasswordresettoken.New(
			logger,
			userRepository,
			passwordResetter,
			passwordResetTokenSender,
		),
	)
	resetPassword := resetpassword.New(
		logger,
		userRepository,
		passwordResetter,
		passwordHasher,
	)

	router := chi.NewRouter()

	router.Post("/auth/signup", handlers.NewSignUpWithEmail(signUpWithEmailService, config.IsTestMode).ServeHTTP)
	router.Post("/auth/signup/anonymously", handlers.NewSignUpAnonymously(signUpAnonymouslyService).ServeHTTP)
	router.Post("/auth/activate", handlers.NewActivateUser(activateUserService).ServeHTTP)
	router.Post("/auth/login", handlers.NewLogInWithEmail(logInWithEmailService).ServeHTTP)
	router.Post("/auth/logout", handlers.NewLogOut(logOutService).ServeHTTP)
	router.Post(
		"/auth/password_reset/send",
		handlers.NewSendPasswordResetToken(sendPasswordResetToken, config.IsTestMode).ServeHTTP,
	)
	router.Post("/auth/password_reset", handlers.NewResetPassword(resetPassword).ServeHTTP)

	address := "0.0.0.0:9090"

	srv := &http.Server{
		Handler:           router,
		Addr:              address,
		ReadTimeout:       5 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      5 * time.Second,
		IdleTimeout:       5 * time.Second,
	}

	logger.Info(context.Background(), "Server has started.", dl.Entry("address", address))
	srv.ListenAndServe()
}
