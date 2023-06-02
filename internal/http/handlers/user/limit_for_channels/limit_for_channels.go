package me

import (
	"errors"
	"net/http"
	e "remindme/internal/core/domain/errors"
	"remindme/internal/core/domain/user"
	"remindme/internal/core/services"
	service "remindme/internal/core/services/get_limit_for_channels"
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
	Email    *response.Limit `json:"email"`
	Telegram *response.Limit `json:"telegram"`
}

func (h *Handler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	result, err := h.service.Run(
		r.Context(),
		service.Input{},
	)
	if errors.Is(err, user.ErrUserDoesNotExist) {
		response.RenderUnauthorized(rw)
		return
	}
	if err != nil {
		response.RenderInternalError(rw)
		return
	}

	res := Result{}
	if result.Email.IsPresent {
		emailLimit := &response.Limit{}
		emailLimit.FromDomain(result.Email.Value)
		res.Email = emailLimit
	}
	if result.Telegram.IsPresent {
		telegramLimit := &response.Limit{}
		telegramLimit.FromDomain(result.Telegram.Value)
		res.Telegram = telegramLimit
	}
	response.Render(rw, res, http.StatusOK)
}
