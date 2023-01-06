package createreminder

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"remindme/internal/core/domain/channel"
	c "remindme/internal/core/domain/common"
	e "remindme/internal/core/domain/errors"
	"remindme/internal/core/domain/reminder"
	"remindme/internal/core/domain/user"
	"remindme/internal/core/services"
	service "remindme/internal/core/services/create_reminder"
	"remindme/internal/http/handlers/response"
	"time"

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
	At         time.Time `json:"at"`
	Every      *string   `json:"every"`
	ChannelIDs []int64   `json:"channel_ids"`
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
		validation.Field(&i.At, validation.Required),
		validation.Field(&i.Every, validation.Length(0, 64)),
		validation.Field(&i.ChannelIDs, validation.Required, validation.Length(1, 5)),
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

	var every c.Optional[reminder.Every]
	if input.Every != nil {
		e, err := reminder.ParseEvery(*input.Every)
		if err != nil {
			response.RenderError(rw, err.Error(), http.StatusBadRequest)
			return
		}
		every = c.NewOptional(e, true)
	}
	channelIDs := make([]channel.ID, len(input.ChannelIDs))
	for ix, channelID := range input.ChannelIDs {
		channelIDs[ix] = channel.ID(channelID)
	}

	result, err := h.service.Run(
		r.Context(),
		service.Input{
			At:         input.At.UTC(),
			Every:      every,
			ChannelIDs: reminder.NewChannelIDs(channelIDs...),
		},
	)
	if err != nil {
		switch {
		case errors.Is(err, user.ErrUserDoesNotExist):
			response.RenderUnauthorized(rw)
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
