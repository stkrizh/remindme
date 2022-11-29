package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"remindme/internal/domain/services"
	"remindme/internal/domain/user"
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

func (s *SignUpWithEmail) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	input := SignUpWithEmailInput{}
	err := input.FromJSON(r.Body)
	if err != nil {
		renderErrorResponse(rw, "Invalid request data.", http.StatusBadRequest)
		return
	}
	_, err = s.service.Run(
		r.Context(),
		services.SignUpWithEmailInput{Email: user.Email(input.Email), Password: user.RawPassword(input.Password)},
	)
	if err == nil {
		renderResponse(rw, "", http.StatusCreated)
	}

	var errEmailAlreadyExists *user.EmailAlreadyExistsError
	if errors.As(err, &errEmailAlreadyExists) {
		renderErrorResponse(rw, "Email already exists.", http.StatusUnprocessableEntity)
		return
	}

	if err != nil {
		renderErrorResponse(rw, "Internal error.", http.StatusInternalServerError)
		return
	}
}
