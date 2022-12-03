package http

import (
	"context"
	"net/http"
	"remindme/internal/config"
	uow "remindme/internal/db/unit_of_work"
	dl "remindme/internal/domain/logging"
	sendactivationemail "remindme/internal/domain/services/send_activation_email"
	signupwithemail "remindme/internal/domain/services/sign_up_with_email"
	"remindme/internal/domain/user"
	"remindme/internal/http/handlers"
	"remindme/internal/implementations/activation"
	"remindme/internal/implementations/logging"
	passwordhasher "remindme/internal/implementations/password_hasher"
	"time"

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

	logger := logging.NewZapLogger()
	defer logger.Sync()

	unitOfWork := uow.NewPgxUnitOfWork(db)

	passwordHasher := passwordhasher.NewBcrypt(config.Secret, config.BcryptHasherCost)
	activationTokenGenerator := activation.NewTokenGenerator()
	activationTokenSender := user.NewFakeActivationTokenSender()

	signUpWithEmailService := sendactivationemail.New(
		logger,
		activationTokenSender,
		signupwithemail.New(
			logger,
			unitOfWork,
			passwordHasher,
			activationTokenGenerator,
			func() time.Time { return time.Now().UTC() },
		),
	)

	router := mux.NewRouter()
	router.Handle("/auth/signup", handlers.NewSignUpWithEmail(signUpWithEmailService))

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
