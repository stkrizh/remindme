package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"remindme/internal/core/domain/user"
	"remindme/internal/core/services"
	activateuser "remindme/internal/core/services/activate_user"

	validation "github.com/go-ozzo/ozzo-validation"
)

type ActivateUser struct {
	service services.Service[activateuser.Input, activateuser.Result]
}

func NewActivateUser(
	service services.Service[activateuser.Input, activateuser.Result],
) *ActivateUser {
	return &ActivateUser{service: service}
}

type ActivateUserInput struct {
	Token string `json:"token"`
}

func (i *ActivateUserInput) FromJSON(r io.Reader) error {
	e := json.NewDecoder(r)
	return e.Decode(i)
}

func (i ActivateUserInput) Validate() error {
	return validation.ValidateStruct(&i,
		validation.Field(&i.Token, validation.Required, validation.Length(0, 128)),
	)
}

func (s *ActivateUser) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	input := ActivateUserInput{}
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
		activateuser.Input{ActivationToken: user.ActivationToken(input.Token)},
	)
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
