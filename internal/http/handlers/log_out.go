package handlers

import (
	"errors"
	"net/http"
	"remindme/internal/domain/services"
	logout "remindme/internal/domain/services/log_out"
	"remindme/internal/domain/user"
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
		renderErrorResponse(rw, "invalid authentication token", http.StatusUnauthorized)
		return
	}
	_, err := s.service.Run(
		r.Context(),
		logout.Input{Token: user.SessionToken(token)},
	)
	if errors.Is(err, user.ErrSessionDoesNotExist) {
		renderErrorResponse(rw, "invalid authentication token", http.StatusUnauthorized)
		return
	}
	if err != nil {
		renderErrorResponse(rw, "internal error", http.StatusInternalServerError)
		return
	}
	renderResponse(rw, struct{}{}, http.StatusOK)
}
