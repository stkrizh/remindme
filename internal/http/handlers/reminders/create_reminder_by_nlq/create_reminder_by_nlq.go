package createreminderbynlq

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	e "remindme/internal/core/domain/errors"
	"remindme/internal/core/domain/reminder"
	"remindme/internal/core/domain/user"
	"remindme/internal/core/services"
	createreminderservice "remindme/internal/core/services/create_reminder"
	service "remindme/internal/core/services/create_reminder_by_nlq"
	"remindme/internal/http/handlers/response"

	validation "github.com/go-ozzo/ozzo-validation"
)

type Handler struct {
	service services.Service[service.Input, createreminderservice.Result]
}

func New(
	service services.Service[service.Input, createreminderservice.Result],
) *Handler {
	if service == nil {
		panic(e.NewNilArgumentError("service"))
	}
	return &Handler{service: service}
}

type Input struct {
	Query string `json:"query"`
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
		validation.Field(&i.Query, validation.Required, validation.Length(0, 128)),
	)
}

func (h *Handler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	input := Input{}
	if err := input.FromJSON(r.Body); err != nil {
		response.RenderError(rw, "invalid request data", http.StatusBadRequest)
		return
	}
	if err := input.Validate(); err != nil {
		response.Render(rw, err, http.StatusBadRequest)
		return
	}

	result, err := h.service.Run(
		r.Context(),
		service.Input{Query: input.Query},
	)
	if err != nil {
		switch {
		case errors.Is(err, user.ErrUserDoesNotExist):
			response.RenderUnauthorized(rw)
		case errors.Is(err, reminder.ErrNaturalQueryParsing):
			response.RenderError(rw, reminder.ErrNaturalQueryParsing.Error(), http.StatusUnprocessableEntity)
		case isExpectedError(err):
			response.RenderError(rw, err.Error(), http.StatusUnprocessableEntity)
		default:
			response.RenderInternalError(rw)
		}
		return
	}

	reminder := response.ReminderWithChannels{}
	reminder.FromDomainType(result.Reminder)
	response.Render(rw, Result{Reminder: reminder}, http.StatusCreated)
}

func isExpectedError(err error) bool {
	return (errors.Is(err, reminder.ErrReminderAtTimeIsNotUTC) ||
		errors.Is(err, reminder.ErrReminderTooEarly) ||
		errors.Is(err, reminder.ErrReminderTooLate) ||
		errors.Is(err, reminder.ErrInvalidEvery) ||
		errors.Is(err, reminder.ErrReminderChannelsNotSet) ||
		errors.Is(err, reminder.ErrReminderChannelsNotValid) ||
		errors.Is(err, reminder.ErrReminderChannelsNotVerified) ||
		errors.Is(err, user.ErrLimitActiveReminderCountExceeded) ||
		errors.Is(err, user.ErrLimitSentReminderCountExceeded) ||
		errors.Is(err, user.ErrLimitReminderEveryPerDayCountExceeded))
}
