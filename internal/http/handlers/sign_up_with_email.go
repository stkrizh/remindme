package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"remindme/internal/core/domain/user"
	"remindme/internal/core/services"
	signupwithemail "remindme/internal/core/services/sign_up_with_email"

	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/is"
)

type SignUpWithEmail struct {
	service    services.Service[signupwithemail.Input, signupwithemail.Result]
	isTestMode bool
}

func NewSignUpWithEmail(
	service services.Service[signupwithemail.Input, signupwithemail.Result],
	isTestMode bool,
) *SignUpWithEmail {
	return &SignUpWithEmail{service: service, isTestMode: isTestMode}
}

type SignUpWithEmailInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (i *SignUpWithEmailInput) FromJSON(r io.Reader) error {
	e := json.NewDecoder(r)
	return e.Decode(i)
}

func (i SignUpWithEmailInput) Validate() error {
	return validation.ValidateStruct(&i,
		validation.Field(&i.Email, validation.Required, is.Email, validation.Length(0, 512)),
		validation.Field(&i.Password, validation.Required, validation.Length(6, 256)),
	)
}

func (s *SignUpWithEmail) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	input := SignUpWithEmailInput{}
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
		signupwithemail.Input{Email: user.NewEmail(input.Email), Password: user.RawPassword(input.Password)},
	)
	if errors.Is(err, context.Canceled) {
		return
	}
	if errors.Is(err, user.ErrEmailAlreadyExists) {
		renderErrorResponse(rw, "email already exists", http.StatusUnprocessableEntity)
		return
	}
	if err != nil {
		renderErrorResponse(rw, "internal error", http.StatusInternalServerError)
		return
	}

	if s.isTestMode {
		rw.Header().Set("x-test-activation-token", string(result.User.ActivationToken.Value))
	}
	renderResponse(rw, struct{}{}, http.StatusCreated)
}
