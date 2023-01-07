package listuserreminders

import (
	"errors"
	"fmt"
	"net/http"
	c "remindme/internal/core/domain/common"
	e "remindme/internal/core/domain/errors"
	"remindme/internal/core/domain/reminder"
	"remindme/internal/core/domain/user"
	"remindme/internal/core/services"
	service "remindme/internal/core/services/list_user_reminders"
	"remindme/internal/http/handlers/response"
	"strconv"
	"strings"
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
	Reminders  []response.ReminderWithChannels `json:"reminders"`
	TotalCount uint                            `json:"total_count"`
}

func (h *Handler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	raw_status_in := r.URL.Query().Get("status_in")
	status_in, err := parseStatusIn(raw_status_in)
	if err != nil {
		response.RenderError(rw, "invalid status_in query parameter", http.StatusBadRequest)
		return
	}

	raw_order_by := r.URL.Query().Get("order_by")
	order_by, err := parseOrderBy(raw_order_by)
	if err != nil {
		response.RenderError(rw, "invalid order_by query parameter", http.StatusBadRequest)
		return
	}

	raw_limit := r.URL.Query().Get("limit")
	limit, err := parseLimit(raw_limit)
	if err != nil {
		response.RenderError(rw, "invalid limit query parameter", http.StatusBadRequest)
		return
	}

	raw_offset := r.URL.Query().Get("offset")
	offset, err := parseOffset(raw_offset)
	if err != nil {
		response.RenderError(rw, "invalid offset query parameter", http.StatusBadRequest)
		return
	}

	input := service.Input{
		StatusIn: status_in,
		OrderBy:  order_by,
		Limit:    limit,
		Offset:   offset,
	}
	result, err := h.service.Run(r.Context(), input)
	if err != nil {
		switch {
		case errors.Is(err, user.ErrUserDoesNotExist):
			response.RenderUnauthorized(rw)
		default:
			response.RenderInternalError(rw)
		}
		return
	}

	respReminders := make([]response.ReminderWithChannels, 0, len(result.Reminders))
	for _, reminder := range result.Reminders {
		respReminder := response.ReminderWithChannels{}
		respReminder.FromDomainType(reminder)
		respReminders = append(respReminders, respReminder)
	}
	response.Render(rw, Result{Reminders: respReminders, TotalCount: result.TotalCount}, http.StatusOK)
}

func parseStatusIn(raw string) (result c.Optional[[]reminder.Status], err error) {
	if raw == "" {
		return result, nil
	}
	raw_statuses := strings.SplitN(raw, ",", 6)
	statuses := make([]reminder.Status, 0, len(raw_statuses))
	for _, raw_status := range raw_statuses {
		status, err := reminder.ParseStatus(raw_status)
		if err != nil {
			return result, err
		}
		statuses = append(statuses, status)
	}

	result.IsPresent = true
	result.Value = statuses
	return result, err
}

func parseOrderBy(raw string) (orderBy reminder.OrderBy, err error) {
	if raw == "" {
		return orderBy, nil
	}
	orderBy, err = reminder.ParseOrderBy(raw)
	return orderBy, err
}

func parseLimit(raw string) (limit c.Optional[uint], err error) {
	if raw == "" {
		return limit, nil
	}
	l, err := strconv.ParseUint(raw, 10, 32)
	if err != nil {
		return limit, err
	}
	if l > service.DEFAULT_LIMIT {
		return limit, fmt.Errorf("limit must be less than or equal to %v", service.DEFAULT_LIMIT)
	}
	limit.IsPresent = true
	limit.Value = uint(l)
	return limit, nil
}

func parseOffset(raw string) (offset uint, err error) {
	if raw == "" {
		return offset, nil
	}
	o, err := strconv.ParseUint(raw, 10, 32)
	if err != nil {
		return offset, err
	}
	return uint(o), nil
}
