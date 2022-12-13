package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	ratelimiter "remindme/internal/core/domain/rate_limiter"
	"remindme/internal/core/domain/user"
	"remindme/internal/core/services"
	service "remindme/internal/core/services/send_password_reset_token"

	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/is"
)

type SendPasswordResetToken struct {
	service    services.Service[service.Input, service.Result]
	isTestMode bool
}

func NewSendPasswordResetToken(
	service services.Service[service.Input, service.Result],
	isTestMode bool,
) *SendPasswordResetToken {
	return &SendPasswordResetToken{service: service, isTestMode: isTestMode}
}

type SendPasswordResetTokenInput struct {
	Email string `json:"email"`
}

func (i *SendPasswordResetTokenInput) FromJSON(r io.Reader) error {
	e := json.NewDecoder(r)
	return e.Decode(i)
}

func (i SendPasswordResetTokenInput) Validate() error {
	return validation.ValidateStruct(&i,
		validation.Field(&i.Email, validation.Required, is.Email, validation.Length(0, 512)),
	)
}

func (s *SendPasswordResetToken) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	input := SendPasswordResetTokenInput{}
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
		service.Input{Email: user.NewEmail(input.Email)},
	)
	if errors.Is(err, ratelimiter.ErrRateLimitExceeded) {
		renderErrorResponse(rw, "rate limit exceeded", http.StatusTooManyRequests)
		return
	}
	if errors.Is(err, user.ErrUserDoesNotExist) {
		renderErrorResponse(rw, "user does not exist", http.StatusUnprocessableEntity)
		return
	}
	if err != nil {
		renderErrorResponse(rw, "internal error", http.StatusInternalServerError)
		return
	}

	if s.isTestMode {
		rw.Header().Set("x-test-password-reset-token", string(result.Token))
	}
	renderResponse(rw, struct{}{}, http.StatusOK)
}
