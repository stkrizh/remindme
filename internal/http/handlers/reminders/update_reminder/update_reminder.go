package updatereminder

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	c "remindme/internal/core/domain/common"
	e "remindme/internal/core/domain/errors"
	"remindme/internal/core/domain/reminder"
	"remindme/internal/core/domain/user"
	"remindme/internal/core/services"
	service "remindme/internal/core/services/update_reminder"
	"remindme/internal/http/handlers/response"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	validation "github.com/go-ozzo/ozzo-validation"
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

type Input struct {
	At            *time.Time `json:"at"`
	DoEveryUpdate bool       `json:"do_every_update"`
	Every         *string    `json:"every"`
	Body          *string    `json:"body"`
}

type Result struct {
	Reminder response.ReminderWithChannels `json:"reminder"`
}

func (i *Input) FromJSON(r io.Reader) error {
	e := json.NewDecoder(r)
	return e.Decode(i)
}

func (i Input) Validate() error {
	return validation.ValidateStruct(&i,
		validation.Field(&i.Every, validation.Length(0, 64)),
		validation.Field(&i.Body, validation.Length(0, reminder.MAX_BODY_LEN)),
	)
}

func (h *Handler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	rawReminderID := chi.URLParam(r, "reminderID")
	reminderID, err := strconv.ParseInt(rawReminderID, 10, 64)
	if err != nil {
		response.RenderError(rw, "invalid reminder ID", http.StatusBadRequest)
		return
	}

	input := Input{}
	if err := input.FromJSON(r.Body); err != nil {
		response.RenderError(rw, "invalid request data", http.StatusBadRequest)
		return
	}
	if err := input.Validate(); err != nil {
		response.Render(rw, err, http.StatusBadRequest)
		return
	}

	var at time.Time
	var doAtUpdate bool
	if input.At != nil {
		doAtUpdate = true
		at = (*input.At).UTC()
	}
	var every c.Optional[reminder.Every]
	if input.DoEveryUpdate && input.Every != nil {
		e, err := reminder.ParseEvery(*input.Every)
		if err != nil {
			response.RenderError(rw, err.Error(), http.StatusBadRequest)
			return
		}
		every = c.NewOptional(e, true)
	}
	var body string
	var doBodyUpdate bool
	if input.Body != nil {
		doBodyUpdate = true
		body = *input.Body
	}

	result, err := h.service.Run(
		r.Context(),
		service.Input{
			ReminderID:    reminder.ID(reminderID),
			DoAtUpdate:    doAtUpdate,
			At:            at,
			DoEveryUpdate: input.DoEveryUpdate,
			Every:         every,
			DoBodyUpdate:  doBodyUpdate,
			Body:          body,
		},
	)
	if err != nil {
		switch {
		case errors.Is(err, user.ErrUserDoesNotExist):
			response.RenderUnauthorized(rw)
		case errors.Is(err, reminder.ErrReminderDoesNotExist):
			response.RenderError(rw, err.Error(), http.StatusNotFound)
		case errors.Is(err, reminder.ErrReminderPermission):
			response.RenderError(rw, err.Error(), http.StatusForbidden)
		case isExpectedError(err):
			response.RenderError(rw, err.Error(), http.StatusUnprocessableEntity)
		default:
			response.RenderInternalError(rw)
		}
		return
	}

	reminder := response.ReminderWithChannels{}
	reminder.FromDomainType(result.Reminder)
	response.Render(rw, Result{Reminder: reminder}, http.StatusOK)
}

func isExpectedError(err error) bool {
	return (errors.Is(err, reminder.ErrReminderAtTimeIsNotUTC) ||
		errors.Is(err, reminder.ErrReminderTooEarly) ||
		errors.Is(err, reminder.ErrReminderTooLate) ||
		errors.Is(err, reminder.ErrInvalidEvery) ||
		errors.Is(err, user.ErrLimitReminderEveryPerDayCountExceeded))
}
