package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"remindme/internal/domain/services"
	signinwithemail "remindme/internal/domain/services/sign_in_with_email"
	"remindme/internal/domain/user"

	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/is"
)

type SignInWithEmail struct {
	service services.Service[signinwithemail.Input, signinwithemail.Result]
}

func NewSignInWithEmail(
	service services.Service[signinwithemail.Input, signinwithemail.Result],
) *SignInWithEmail {
	return &SignInWithEmail{service: service}
}

type SignInWithEmailInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type SignInWithEmailResult struct {
	Token string `json:"token"`
}

func (i *SignInWithEmailInput) FromJSON(r io.Reader) error {
	e := json.NewDecoder(r)
	return e.Decode(i)
}

func (i SignInWithEmailInput) Validate() error {
	return validation.ValidateStruct(&i,
		validation.Field(&i.Email, validation.Required, is.Email, validation.Length(0, 512)),
		validation.Field(&i.Password, validation.Required, validation.Length(0, 512)),
	)
}

func (s *SignInWithEmail) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	input := SignInWithEmailInput{}
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
		signinwithemail.Input{Email: user.NewEmail(input.Email), Password: user.RawPassword(input.Password)},
	)
	if errors.Is(err, user.ErrInvalidCredentials) {
		renderErrorResponse(rw, "invalid credentials", http.StatusUnauthorized)
		return
	}
	if err != nil {
		renderErrorResponse(rw, "internal error", http.StatusInternalServerError)
		return
	}

	renderResponse(rw, SignInWithEmailResult{Token: string(result.Token)}, http.StatusOK)
}
