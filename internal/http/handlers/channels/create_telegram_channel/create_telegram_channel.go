package createtelegramchannel

import (
	"errors"
	"net/http"
	"remindme/internal/core/domain/channel"
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
	return &Handler{service: service, defaultTelegramBot: defaultTelegramBot}
}

type Result struct {
	ChannelID         int64  `json:"channel_id"`
	VerificationToken string `json:"token"`
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

	response.Render(
		rw,
		Result{
			ChannelID:         int64(result.Channel.ID),
			VerificationToken: string(result.VerificationToken),
		},
		http.StatusCreated,
	)
}
