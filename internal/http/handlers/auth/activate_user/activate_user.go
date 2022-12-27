package activateuser

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	e "remindme/internal/core/domain/errors"
	"remindme/internal/core/domain/user"
	"remindme/internal/core/services"
	activateuser "remindme/internal/core/services/activate_user"
	"remindme/internal/http/handlers/response"

	validation "github.com/go-ozzo/ozzo-validation"
)

type Handler struct {
	service services.Service[activateuser.Input, activateuser.Result]
}

func New(
	service services.Service[activateuser.Input, activateuser.Result],
) *Handler {
	if service == nil {
		panic(e.NewNilArgumentError("service"))
	}
	return &Handler{service: service}
}

type Input struct {
	Token string `json:"token"`
}

func (i *Input) FromJSON(r io.Reader) error {
	e := json.NewDecoder(r)
	return e.Decode(i)
}

func (i Input) Validate() error {
	return validation.ValidateStruct(&i,
		validation.Field(&i.Token, validation.Required, validation.Length(0, 128)),
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
		activateuser.Input{ActivationToken: user.ActivationToken(input.Token)},
	)
	if errors.Is(err, user.ErrUserDoesNotExist) {
		response.RenderError(rw, "user does not exist", http.StatusUnprocessableEntity)
		return
	}
	if err != nil {
		response.RenderInternalError(rw)
		return
	}

	response.Render(rw, struct{}{}, http.StatusOK)
}
