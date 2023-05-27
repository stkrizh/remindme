package me

import (
	"errors"
	"net/http"
	e "remindme/internal/core/domain/errors"
	"remindme/internal/core/domain/user"
	"remindme/internal/core/services"
	service "remindme/internal/core/services/get_user_limits"
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
	Limits response.Limits `json:"limits"`
	Values response.Limits `json:"values"`
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

	limits := response.Limits{}
	limits.FromDomainLimits(result.Limits)
	values := response.Limits{}
	values.FromDomainLimits(result.Values)
	response.Render(rw, Result{Limits: limits, Values: values}, http.StatusOK)
}
