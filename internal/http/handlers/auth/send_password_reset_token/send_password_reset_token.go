package sendpasswordresettoken

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	c "remindme/internal/core/domain/common"
	e "remindme/internal/core/domain/errors"
	ratelimiter "remindme/internal/core/domain/rate_limiter"
	"remindme/internal/core/domain/user"
	"remindme/internal/core/services"
	"remindme/internal/core/services/captcha"
	service "remindme/internal/core/services/send_password_reset_token"
	"remindme/internal/http/handlers/response"

	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/is"
)

type Handler struct {
	service    services.Service[service.Input, service.Result]
	isTestMode bool
}

func New(
	service services.Service[service.Input, service.Result],
	isTestMode bool,
) *Handler {
	if service == nil {
		panic(e.NewNilArgumentError("service"))
	}
	return &Handler{service: service, isTestMode: isTestMode}
}

type Input struct {
	Email string `json:"email"`
}

func (i *Input) FromJSON(r io.Reader) error {
	e := json.NewDecoder(r)
	return e.Decode(i)
}

func (i Input) Validate() error {
	return validation.ValidateStruct(&i,
		validation.Field(&i.Email, validation.Required, is.Email, validation.Length(0, 512)),
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
		service.Input{Email: c.NewEmail(input.Email)},
	)
	if err != nil {
		switch {
		case errors.Is(err, captcha.ErrInvalidCaptcha):
			response.RenderError(rw, err.Error(), http.StatusUnprocessableEntity)
		case errors.Is(err, ratelimiter.ErrRateLimitExceeded):
			response.RenderError(rw, "rate limit exceeded", http.StatusTooManyRequests)
		case errors.Is(err, user.ErrUserDoesNotExist):
			response.RenderError(rw, "user does not exist", http.StatusUnprocessableEntity)
		default:
			response.RenderInternalError(rw)
		}
		return
	}

	if h.isTestMode {
		rw.Header().Set("x-test-password-reset-token", string(result.Token))
	}
	response.Render(rw, struct{}{}, http.StatusOK)
}
