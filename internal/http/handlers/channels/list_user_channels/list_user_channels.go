package listuserchannels

import (
	"errors"
	"net/http"
	e "remindme/internal/core/domain/errors"
	"remindme/internal/core/domain/user"
	"remindme/internal/core/services"
	service "remindme/internal/core/services/list_user_channels"
	"remindme/internal/http/handlers/response"
)

type Handler struct {
	service services.Service[service.Input, service.Result]
}

func New(
	service services.Service[service.Input, service.Result],
) *Handler {
	if service == nil {
		panic(e.NewNilArgumentError("service"))
	}
	return &Handler{service: service}
}

type Result struct {
	Channels []response.Channel `json:"channels"`
}

func (h *Handler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	result, err := h.service.Run(r.Context(), service.Input{})
	if err != nil {
		switch {
		case errors.Is(err, user.ErrUserDoesNotExist):
			response.RenderUnauthorized(rw)
		default:
			response.RenderInternalError(rw)
		}
		return
	}

	respChannels := make([]response.Channel, len(result.Channels))
	for ix, channel := range result.Channels {
		respChannel := response.Channel{}
		respChannel.FromDomainChannel(channel)
		respChannels[ix] = respChannel
	}
	response.Render(rw, Result{Channels: respChannels}, http.StatusOK)
}
