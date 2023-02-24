package updateuser

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	e "remindme/internal/core/domain/errors"
	"remindme/internal/core/domain/user"
	"remindme/internal/core/services"
	service "remindme/internal/core/services/update_user"
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
	TimeZone *string `json:"timezone"`
}

type Result struct {
	User response.User `json:"user"`
}

func (i *Input) FromJSON(r io.Reader) error {
	e := json.NewDecoder(r)
	return e.Decode(i)
}

func (i Input) Validate() error {
	return validation.ValidateStruct(&i,
		validation.Field(&i.TimeZone, validation.Length(0, 64)),
	)
}

func (h *Handler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	input := Input{}
	if err := input.FromJSON(r.Body); err != nil {
		response.RenderError(rw, "invalid request data", http.StatusBadRequest)
		return
	}

	var doTimeZoneUpdate bool
	var tz *time.Location
	if input.TimeZone != nil {
		doTimeZoneUpdate = true
		parsedTimeZone, err := time.LoadLocation(*input.TimeZone)
		if err != nil {
			response.RenderError(rw, "invalid timezone", http.StatusBadRequest)
			return
		}
		tz = parsedTimeZone
	}

	result, err := h.service.Run(
		r.Context(),
		service.Input{
			DoTimeZoneUpdate: doTimeZoneUpdate,
			TimeZone:         tz,
		},
	)
	if err != nil {
		switch {
		case errors.Is(err, user.ErrUserDoesNotExist):
			response.RenderUnauthorized(rw)
		default:
			response.RenderInternalError(rw)
		}
		return
	}

	user := response.User{}
	user.FromDomainUser(result.User)
	response.Render(rw, Result{User: user}, http.StatusOK)
}
