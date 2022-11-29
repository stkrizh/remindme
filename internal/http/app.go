package http

import (
	"context"
	"net/http"
	"os"
	uow "remindme/internal/db/unit_of_work"
	sendactivationemail "remindme/internal/domain/services/send_activation_email"
	signupwithemail "remindme/internal/domain/services/sign_up_with_email"
	"remindme/internal/domain/user"
	"remindme/internal/http/handlers"
	"remindme/internal/logging"
	"time"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v4/pgxpool"
)

func StartApp() {
	dbConnString := os.Getenv("POSTGRESQL_URL")
	if dbConnString == "" {
		panic("POSTGRESQL_URL must be set.")
	}
	db, err := pgxpool.Connect(context.Background(), dbConnString)
	if err != nil {
		panic("Could not connect to the database.")
	}
	defer db.Close()

	logger := logging.NewZapLogger()
	defer logger.Sync()
	unitOfWork := uow.NewPgxUnitOfWork(db)

	passwordHasher := user.NewFakePasswordHasher()
	activationTokenGenerator := user.NewFakeActivationTokenGenerator("test")
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
	// rest.NewTaskHandler(svc).Register(r)

	address := "0.0.0.0:9090"

	srv := &http.Server{
		Handler:           router,
		Addr:              address,
		ReadTimeout:       5 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      5 * time.Second,
		IdleTimeout:       5 * time.Second,
	}

	srv.ListenAndServe()
}
