package logout

import (
	"errors"
	"net/http"
	"remindme/internal/core/domain/user"
	"remindme/internal/core/services"
	logout "remindme/internal/core/services/log_out"
	"remindme/internal/http/handlers/auth"
	"remindme/internal/http/handlers/response"
)

type Handler struct {
	service services.Service[logout.Input, logout.Result]
}

func New(
	service services.Service[logout.Input, logout.Result],
) *Handler {
	return &Handler{service: service}
}

func (h *Handler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	token, ok := auth.ParseToken(r)
	if !ok {
		response.RenderUnauthorized(rw)
		return
	}
	_, err := h.service.Run(
		r.Context(),
		logout.Input{Token: user.SessionToken(token)},
	)
	if errors.Is(err, user.ErrSessionDoesNotExist) {
		response.RenderUnauthorized(rw)
		return
	}
	if err != nil {
		response.RenderInternalError(rw)
		return
	}
	response.Render(rw, struct{}{}, http.StatusOK)
}
