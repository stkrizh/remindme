package signupwithemail

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"remindme/internal/core/domain/user"
	"remindme/internal/core/services"
	signupwithemail "remindme/internal/core/services/sign_up_with_email"
	"remindme/internal/http/handlers/response"

	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/is"
)

type Handler struct {
	service    services.Service[signupwithemail.Input, signupwithemail.Result]
	isTestMode bool
}

func New(
	service services.Service[signupwithemail.Input, signupwithemail.Result],
	isTestMode bool,
) *Handler {
	return &Handler{service: service, isTestMode: isTestMode}
}

type Input struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (i *Input) FromJSON(r io.Reader) error {
	e := json.NewDecoder(r)
	return e.Decode(i)
}

func (i Input) Validate() error {
	return validation.ValidateStruct(&i,
		validation.Field(&i.Email, validation.Required, is.Email, validation.Length(0, 512)),
		validation.Field(&i.Password, validation.Required, validation.Length(6, 256)),
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

	result, err := h.service.Run(
		r.Context(),
		signupwithemail.Input{Email: user.NewEmail(input.Email), Password: user.RawPassword(input.Password)},
	)
	if errors.Is(err, user.ErrEmailAlreadyExists) {
		response.RenderError(rw, "email already exists", http.StatusUnprocessableEntity)
		return
	}
	if err != nil {
		response.RenderInternalError(rw)
		return
	}

	if h.isTestMode {
		rw.Header().Set("x-test-activation-token", string(result.User.ActivationToken.Value))
	}
	response.Render(rw, struct{}{}, http.StatusCreated)
}
