package signupwithemail

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	c "remindme/internal/core/domain/common"
	e "remindme/internal/core/domain/errors"
	"remindme/internal/core/domain/user"
	"remindme/internal/core/services"
	"remindme/internal/core/services/captcha"
	signupwithemail "remindme/internal/core/services/sign_up_with_email"
	"remindme/internal/http/handlers/response"
	"time"

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
	if service == nil {
		panic(e.NewNilArgumentError("service"))
	}
	return &Handler{service: service, isTestMode: isTestMode}
}

type Input struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	TimeZone string `json:"timezone"`
}

func (i *Input) FromJSON(r io.Reader) error {
	e := json.NewDecoder(r)
	return e.Decode(i)
}

func (i Input) Validate() error {
	return validation.ValidateStruct(&i,
		validation.Field(&i.Email, validation.Required, is.Email, validation.Length(0, 512)),
		validation.Field(&i.Password, validation.Required, validation.Length(8, 256)),
		validation.Field(&i.TimeZone, validation.Required, validation.Length(1, 64)),
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
	tz, err := time.LoadLocation(input.TimeZone)
	if err != nil {
		response.RenderError(rw, "invalid timezone", http.StatusBadRequest)
		return
	}

	result, err := h.service.Run(
		r.Context(),
		signupwithemail.Input{
			Email:    c.NewEmail(input.Email),
			Password: user.RawPassword(input.Password),
			TimeZone: tz,
		},
	)

	if err != nil {
		switch {
		case errors.Is(err, captcha.ErrInvalidCaptcha):
			response.RenderError(rw, err.Error(), http.StatusUnprocessableEntity)
		case errors.Is(err, user.ErrEmailAlreadyExists):
			response.RenderError(rw, "email already exists", http.StatusUnprocessableEntity)
		default:
			response.RenderInternalError(rw)
		}
		return
	}

	if h.isTestMode {
		rw.Header().Set("x-test-activation-token", string(result.User.ActivationToken.Value))
	}
	response.Render(rw, struct{}{}, http.StatusCreated)
}
