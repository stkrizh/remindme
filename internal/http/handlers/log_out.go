package handlers

import (
	"errors"
	"net/http"
	"remindme/internal/core/domain/user"
	"remindme/internal/core/services"
	logout "remindme/internal/core/services/log_out"
)

type LogOut struct {
	service services.Service[logout.Input, logout.Result]
}

func NewLogOut(
	service services.Service[logout.Input, logout.Result],
) *LogOut {
	return &LogOut{service: service}
}

func (s *LogOut) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	token, ok := GetAuthToken(r)
	if !ok {
		renderUnauthorizedResponse(rw)
		return
	}
	_, err := s.service.Run(
		r.Context(),
		logout.Input{Token: user.SessionToken(token)},
	)
	if errors.Is(err, user.ErrSessionDoesNotExist) {
		renderUnauthorizedResponse(rw)
		return
	}
	if err != nil {
		renderErrorResponse(rw, "internal error", http.StatusInternalServerError)
		return
	}
	renderResponse(rw, struct{}{}, http.StatusOK)
}
