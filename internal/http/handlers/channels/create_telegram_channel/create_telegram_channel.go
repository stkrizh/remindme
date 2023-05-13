package createtelegramchannel

import (
	"errors"
	"net/http"
	"remindme/internal/core/domain/channel"
	e "remindme/internal/core/domain/errors"
	"remindme/internal/core/domain/user"
	"remindme/internal/core/services"
	service "remindme/internal/core/services/create_telegram_channel"
	"remindme/internal/http/handlers/response"
)

type Handler struct {
	service            services.Service[service.Input, service.Result]
	defaultTelegramBot channel.TelegramBot
}

func New(
	service services.Service[service.Input, service.Result],
	defaultTelegramBot channel.TelegramBot,
) *Handler {
	if service == nil {
		panic(e.NewNilArgumentError("service"))
	}
	return &Handler{service: service, defaultTelegramBot: defaultTelegramBot}
}

type Result struct {
	Channel response.Channel `json:"channel"`
}

func (h *Handler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	result, err := h.service.Run(
		r.Context(),
		service.Input{Bot: h.defaultTelegramBot},
	)
	if err != nil {
		switch {
		case errors.Is(err, user.ErrUserDoesNotExist):
			response.RenderUnauthorized(rw)
		case errors.Is(err, user.ErrLimitTelegramChannelCountExceeded):
			response.RenderError(rw, err.Error(), http.StatusUnprocessableEntity)
		default:
			response.RenderInternalError(rw)
		}
		return
	}

	channel := response.Channel{}
	channel.FromDomainChannel(result.Channel)
	response.Render(
		rw,
		Result{
			Channel: channel,
		},
		http.StatusCreated,
	)
}
