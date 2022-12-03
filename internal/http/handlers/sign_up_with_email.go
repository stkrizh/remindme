package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"remindme/internal/domain/services"
	"remindme/internal/domain/user"

	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/is"
)

type SignUpWithEmail struct {
	service services.Service[services.SignUpWithEmailInput, services.SignUpWithEmailResult]
}

func NewSignUpWithEmail(
	service services.Service[services.SignUpWithEmailInput, services.SignUpWithEmailResult],
) *SignUpWithEmail {
	return &SignUpWithEmail{service: service}
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
		validation.Field(&i.Email, validation.Required, is.Email),
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

	_, err := s.service.Run(
		r.Context(),
		services.SignUpWithEmailInput{Email: user.Email(input.Email), Password: user.RawPassword(input.Password)},
	)
	if err == nil {
		renderResponse(rw, struct{}{}, http.StatusCreated)
		return
	}

	var errEmailAlreadyExists *user.EmailAlreadyExistsError
	if errors.As(err, &errEmailAlreadyExists) {
		renderErrorResponse(rw, "email already exists", http.StatusUnprocessableEntity)
		return
	}

	renderErrorResponse(rw, "internal error", http.StatusInternalServerError)
}
