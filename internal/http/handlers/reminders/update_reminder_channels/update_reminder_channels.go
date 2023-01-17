package updatereminderchannels

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"remindme/internal/core/domain/channel"
	e "remindme/internal/core/domain/errors"
	"remindme/internal/core/domain/reminder"
	"remindme/internal/core/domain/user"
	"remindme/internal/core/services"
	service "remindme/internal/core/services/update_reminder_channels"
	"remindme/internal/http/handlers/response"
	"strconv"

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
	ChannelIDs []int64 `json:"channel_ids"`
}

type Result struct {
	ChannelIDs []int64 `json:"channel_ids"`
}

func (i *Input) FromJSON(r io.Reader) error {
	e := json.NewDecoder(r)
	return e.Decode(i)
}

func (i Input) Validate() error {
	return validation.ValidateStruct(&i,
		validation.Field(&i.ChannelIDs, validation.Required, validation.Length(1, reminder.MAX_CHANNEL_COUNT)),
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

	channelIDs := make([]channel.ID, len(input.ChannelIDs))
	for ix, channelID := range input.ChannelIDs {
		channelIDs[ix] = channel.ID(channelID)
	}

	result, err := h.service.Run(
		r.Context(),
		service.Input{
			ReminderID: reminder.ID(reminderID),
			ChannelIDs: reminder.NewChannelIDs(channelIDs...),
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

	resultChannelIDs := make([]int64, 0, len(result.ChannelIDs))
	for _, channelID := range result.ChannelIDs {
		resultChannelIDs = append(resultChannelIDs, int64(channelID))
	}
	response.Render(rw, Result{ChannelIDs: resultChannelIDs}, http.StatusOK)
}

func isExpectedError(err error) bool {
	return (errors.Is(err, reminder.ErrReminderChannelsNotValid) ||
		errors.Is(err, reminder.ErrReminderChannelsNotVerified))
}
