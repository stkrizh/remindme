package verifychannel

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"remindme/internal/core/domain/channel"
	ratelimiter "remindme/internal/core/domain/rate_limiter"
	"remindme/internal/core/domain/user"
	"remindme/internal/core/services"
	service "remindme/internal/core/services/verify_channel"
	"remindme/internal/http/handlers/response"
	"strconv"

	"github.com/go-chi/chi/v5"
	validation "github.com/go-ozzo/ozzo-validation"
)

type Handler struct {
	service services.Service[service.Input, service.Result]
}

func New(
	service services.Service[service.Input, service.Result],
) *Handler {
	return &Handler{service: service}
}

type Input struct {
	Token string `json:"token"`
}

type Result struct {
	Channel response.Channel `json:"channel"`
}

func (i *Input) FromJSON(r io.Reader) error {
	e := json.NewDecoder(r)
	return e.Decode(i)
}

func (i Input) Validate() error {
	return validation.ValidateStruct(&i,
		validation.Field(&i.Token, validation.Required, validation.Length(1, 512)),
	)
}

func (h *Handler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	rawChannelID := chi.URLParam(r, "channelID")
	channelID, err := strconv.ParseInt(rawChannelID, 10, 64)
	if err != nil {
		response.RenderError(rw, "invalid channel ID", http.StatusBadRequest)
		return
	}
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
		service.Input{
			ChannelID:         channel.ID(channelID),
			VerificationToken: channel.VerificationToken(input.Token),
		},
	)
	if err != nil {
		switch {
		case errors.Is(err, user.ErrUserDoesNotExist):
			response.RenderUnauthorized(rw)
		case errors.Is(err, ratelimiter.ErrRateLimitExceeded):
			response.RenderRateLimitExceeded(rw)
		case errors.Is(err, channel.ErrChannelDoesNotExist):
			response.RenderError(rw, err.Error(), http.StatusUnprocessableEntity)
		default:
			response.RenderInternalError(rw)
		}
		return
	}

	respChannel := response.Channel{}
	respChannel.FromDomainChannel(result.Channel)
	response.Render(rw, Result{Channel: respChannel}, http.StatusOK)
}
