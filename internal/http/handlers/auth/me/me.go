package me

import (
	"errors"
	"net/http"
	e "remindme/internal/core/domain/errors"
	"remindme/internal/core/domain/user"
	"remindme/internal/core/services"
	service "remindme/internal/core/services/get_user_by_session_token"
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
	User response.User `json:"user"`
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

	user := response.User{}
	user.FromDomainUser(result.User)
	response.Render(rw, Result{User: user}, http.StatusOK)
}
