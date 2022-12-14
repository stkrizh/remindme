package loginwithemail

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	c "remindme/internal/core/domain/common"
	ratelimiter "remindme/internal/core/domain/rate_limiter"
	"remindme/internal/core/domain/user"
	"remindme/internal/core/services"
	loginwithemail "remindme/internal/core/services/log_in_with_email"
	"remindme/internal/http/handlers/response"

	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/is"
)

type Handler struct {
	service services.Service[loginwithemail.Input, loginwithemail.Result]
}

func New(
	service services.Service[loginwithemail.Input, loginwithemail.Result],
) *Handler {
	return &Handler{service: service}
}

type Input struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type Result struct {
	Token string `json:"token"`
}

func (i *Input) FromJSON(r io.Reader) error {
	e := json.NewDecoder(r)
	return e.Decode(i)
}

func (i Input) Validate() error {
	return validation.ValidateStruct(&i,
		validation.Field(&i.Email, validation.Required, is.Email, validation.Length(0, 512)),
		validation.Field(&i.Password, validation.Required, validation.Length(0, 512)),
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
		loginwithemail.Input{Email: c.NewEmail(input.Email), Password: user.RawPassword(input.Password)},
	)
	if errors.Is(err, ratelimiter.ErrRateLimitExceeded) {
		response.RenderRateLimitExceeded(rw)
		return
	}
	if errors.Is(err, user.ErrInvalidCredentials) {
		response.RenderError(rw, "invalid credentials", http.StatusUnauthorized)
		return
	}
	if errors.Is(err, user.ErrUserIsNotActive) {
		response.RenderError(rw, "user is not active", http.StatusUnprocessableEntity)
		return
	}
	if err != nil {
		response.RenderInternalError(rw)
		return
	}

	response.Render(rw, Result{Token: string(result.Token)}, http.StatusOK)
}
