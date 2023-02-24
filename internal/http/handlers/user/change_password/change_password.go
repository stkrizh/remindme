package changepassword

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	e "remindme/internal/core/domain/errors"
	"remindme/internal/core/domain/user"
	"remindme/internal/core/services"
	changepassword "remindme/internal/core/services/change_password"
	"remindme/internal/http/handlers/response"

	validation "github.com/go-ozzo/ozzo-validation"
)

type Handler struct {
	service services.Service[changepassword.Input, changepassword.Result]
}

func New(
	service services.Service[changepassword.Input, changepassword.Result],
) *Handler {
	if service == nil {
		panic(e.NewNilArgumentError("service"))
	}
	return &Handler{service: service}
}

type Input struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

func (i *Input) FromJSON(r io.Reader) error {
	e := json.NewDecoder(r)
	return e.Decode(i)
}

func (i Input) Validate() error {
	return validation.ValidateStruct(&i,
		validation.Field(&i.CurrentPassword, validation.Required, validation.Length(0, 256)),
		validation.Field(&i.NewPassword, validation.Required, validation.Length(6, 256)),
	)
}

func (h *Handler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	input := Input{}
	if err := input.FromJSON(r.Body); err != nil {
		response.RenderError(rw, "invalid request data", http.StatusBadRequest)
		return
	}
	if err := input.Validate(); err != nil {
		response.Render(rw, err, http.StatusBadRequest)
		return
	}

	_, err := h.service.Run(
		r.Context(),
		changepassword.Input{
			CurrentPassword: user.RawPassword(input.CurrentPassword),
			NewPassword:     user.RawPassword(input.NewPassword),
		},
	)
	if err != nil {
		switch {
		case errors.Is(err, user.ErrUserDoesNotExist):
			response.RenderUnauthorized(rw)
		case errors.Is(err, user.ErrInvalidCredentials):
			response.RenderError(rw, err.Error(), http.StatusUnprocessableEntity)
		default:
			response.RenderInternalError(rw)
		}
		return
	}

	response.Render(rw, struct{}{}, http.StatusOK)
}
