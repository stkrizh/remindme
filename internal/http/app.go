package http

import (
	"context"
	"net/http"
	"remindme/internal/config"
	uow "remindme/internal/db/unit_of_work"
	dbuser "remindme/internal/db/user"
	dl "remindme/internal/domain/logging"
	drl "remindme/internal/domain/rate_limiter"
	signinwithemail "remindme/internal/domain/services/sign_in_with_email"
	signupanonymously "remindme/internal/domain/services/sign_up_anonymously"
	signupwithemail "remindme/internal/domain/services/sign_up_with_email"
	"remindme/internal/domain/user"
	"remindme/internal/http/handlers"
	"remindme/internal/implementations/activation"
	"remindme/internal/implementations/identity"
	"remindme/internal/implementations/logging"
	passwordhasher "remindme/internal/implementations/password_hasher"
	ratelimiter "remindme/internal/implementations/rate_limiter"
	"remindme/internal/implementations/session"
	"time"

	"github.com/go-redis/redis/v9"
	"github.com/gorilla/mux"
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
	signUpAnonymouslyService := signupanonymously.NewWithRateLimiting(
		logger,
		rateLimiter,
		drl.Limit{Value: 3, Interval: drl.Minute},
		signupanonymously.New(
			logger,
			unitOfWork,
			identityGenerator,
			sessionTokenGenerator,
			now,
		),
	)
	signInWithEmailService := signinwithemail.New(
		logger,
		userRepository,
		sessionRepository,
		passwordHasher,
		sessionTokenGenerator,
		now,
	)

	router := mux.NewRouter()
	router.Handle("/auth/signup", handlers.NewSignUpWithEmail(signUpWithEmailService))
	router.Handle("/auth/signup/anonymously", handlers.NewSignUpAnonymously(signUpAnonymouslyService))
	router.Handle("/auth/signin", handlers.NewSignInWithEmail(signInWithEmailService))

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
