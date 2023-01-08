package cancelreminder

import (
	"errors"
	"net/http"
	e "remindme/internal/core/domain/errors"
	"remindme/internal/core/domain/reminder"
	"remindme/internal/core/domain/user"
	"remindme/internal/core/services"
	service "remindme/internal/core/services/cancel_reminder"
	"remindme/internal/http/handlers/response"
	"strconv"

	"github.com/go-chi/chi/v5"
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
	Reminder response.ReminderWithChannels `json:"reminder"`
}

func (h *Handler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	rawReminderID := chi.URLParam(r, "reminderID")
	reminderID, err := strconv.ParseInt(rawReminderID, 10, 64)
	if err != nil {
		response.RenderError(rw, "invalid reminder ID", http.StatusBadRequest)
		return
	}

	result, err := h.service.Run(r.Context(), service.Input{ReminderID: reminder.ID(reminderID)})
	if err != nil {
		switch {
		case errors.Is(err, user.ErrUserDoesNotExist):
			response.RenderUnauthorized(rw)
		case errors.Is(err, reminder.ErrReminderDoesNotExist):
			response.RenderError(rw, err.Error(), http.StatusNotFound)
		case errors.Is(err, reminder.ErrReminderPermission):
			response.RenderError(rw, err.Error(), http.StatusForbidden)
		case errors.Is(err, reminder.ErrReminderNotActive):
			response.RenderError(rw, err.Error(), http.StatusUnprocessableEntity)
		default:
			response.RenderInternalError(rw)
		}
		return
	}

	var reminder response.ReminderWithChannels
	reminder.FromDomainType(result.Reminder)
	response.Render(rw, Result{Reminder: reminder}, http.StatusOK)
}
