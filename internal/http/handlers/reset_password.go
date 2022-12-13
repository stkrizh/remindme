package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"remindme/internal/core/domain/user"
	"remindme/internal/core/services"
	resetpassword "remindme/internal/core/services/reset_password"

	validation "github.com/go-ozzo/ozzo-validation"
)

type ResetPassword struct {
	service services.Service[resetpassword.Input, resetpassword.Result]
}

func NewResetPassword(
	service services.Service[resetpassword.Input, resetpassword.Result],
) *ResetPassword {
	return &ResetPassword{service: service}
}

type ResetPasswordInput struct {
	Token    string `json:"token"`
	Password string `json:"password"`
}

func (i *ResetPasswordInput) FromJSON(r io.Reader) error {
	e := json.NewDecoder(r)
	return e.Decode(i)
}

func (i ResetPasswordInput) Validate() error {
	return validation.ValidateStruct(&i,
		validation.Field(&i.Token, validation.Required, validation.Length(0, 1024)),
		validation.Field(&i.Password, validation.Required, validation.Length(6, 256)),
	)
}

func (s *ResetPassword) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	input := ResetPasswordInput{}
	if err := input.FromJSON(r.Body); err != nil {
		renderErrorResponse(rw, "invalid request data", http.StatusBadRequest)
		return
	}
	if err := input.Validate(); err != nil {
		renderResponse(rw, err, http.StatusBadRequest)
		return
	}

	_, err := s.service.Run(
		r.Context(),
		resetpassword.Input{
			Token:       user.PasswordResetToken(input.Token),
			NewPassword: user.RawPassword(input.Password),
		},
	)
	if errors.Is(err, user.ErrInvalidPasswordFResetToken) {
		renderErrorResponse(rw, "invalid token", http.StatusUnprocessableEntity)
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

	renderResponse(rw, struct{}{}, http.StatusOK)
}
