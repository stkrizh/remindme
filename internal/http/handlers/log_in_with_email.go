package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	ratelimiter "remindme/internal/domain/rate_limiter"
	"remindme/internal/domain/services"
	loginwithemail "remindme/internal/domain/services/log_in_with_email"
	"remindme/internal/domain/user"

	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/is"
)

type LogInWithEmail struct {
	service services.Service[loginwithemail.Input, loginwithemail.Result]
}

func NewLogInWithEmail(
	service services.Service[loginwithemail.Input, loginwithemail.Result],
) *LogInWithEmail {
	return &LogInWithEmail{service: service}
}

type LogInWithEmailInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LogInWithEmailResult struct {
	Token string `json:"token"`
}

func (i *LogInWithEmailInput) FromJSON(r io.Reader) error {
	e := json.NewDecoder(r)
	return e.Decode(i)
}

func (i LogInWithEmailInput) Validate() error {
	return validation.ValidateStruct(&i,
		validation.Field(&i.Email, validation.Required, is.Email, validation.Length(0, 512)),
		validation.Field(&i.Password, validation.Required, validation.Length(0, 512)),
	)
}

func (s *LogInWithEmail) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	input := LogInWithEmailInput{}
	if err := input.FromJSON(r.Body); err != nil {
		renderErrorResponse(rw, "invalid request data", http.StatusBadRequest)
		return
	}
	if err := input.Validate(); err != nil {
		renderResponse(rw, err, http.StatusBadRequest)
		return
	}

	result, err := s.service.Run(
		r.Context(),
		loginwithemail.Input{Email: user.NewEmail(input.Email), Password: user.RawPassword(input.Password)},
	)
	if errors.Is(err, ratelimiter.ErrRateLimitExceeded) {
		renderErrorResponse(rw, "rate limit exceeded", http.StatusTooManyRequests)
		return
	}
	if errors.Is(err, user.ErrInvalidCredentials) {
		renderErrorResponse(rw, "invalid credentials", http.StatusUnauthorized)
		return
	}
	if errors.Is(err, user.ErrUserIsNotActive) {
		renderErrorResponse(rw, "user is not active", http.StatusUnprocessableEntity)
		return
	}
	if err != nil {
		renderErrorResponse(rw, "internal error", http.StatusInternalServerError)
		return
	}

	renderResponse(rw, LogInWithEmailResult{Token: string(result.Token)}, http.StatusOK)
}
