package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/netip"
	e "remindme/internal/core/domain/errors"
	"remindme/internal/core/services"
	signupanonymously "remindme/internal/core/services/sign_up_anonymously"
	"remindme/internal/http/handlers/response"
	"time"

	validation "github.com/go-ozzo/ozzo-validation"
)

const IP_HEADER = "X-Real-IP"

type Handler struct {
	service services.Service[signupanonymously.Input, signupanonymously.Result]
}

func New(
	service services.Service[signupanonymously.Input, signupanonymously.Result],
) *Handler {
	if service == nil {
		panic(e.NewNilArgumentError("service"))
	}
	return &Handler{service: service}
}

type Input struct {
	TimeZone string `json:"timezone"`
}

func (i *Input) FromJSON(r io.Reader) error {
	e := json.NewDecoder(r)
	return e.Decode(i)
}

func (i Input) Validate() error {
	return validation.ValidateStruct(&i,
		validation.Field(&i.TimeZone, validation.Required, validation.Length(1, 64)),
	)
}

type Result struct {
	Identity string `json:"identity"`
	Token    string `json:"token"`
}

func (h *Handler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	ipRaw := r.Header.Get(IP_HEADER)
	if ipRaw == "" {
		response.RenderError(rw, fmt.Sprintf("%s header must be provided", IP_HEADER), http.StatusBadRequest)
		return
	}
	ip, err := netip.ParseAddr(ipRaw)
	if err != nil {
		response.RenderError(rw, fmt.Sprintf("invalid %s header value", IP_HEADER), http.StatusBadRequest)
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
	tz, err := time.LoadLocation(input.TimeZone)
	if err != nil {
		response.RenderError(rw, "invalid timezone", http.StatusBadRequest)
		return
	}

	result, err := h.service.Run(r.Context(), signupanonymously.Input{IP: ip, TimeZone: tz})
	if err != nil {
		response.RenderInternalError(rw)
		return
	}

	response.Render(
		rw,
		Result{
			Identity: string(result.User.Identity.Value),
			Token:    string(result.Token),
		},
		http.StatusCreated,
	)
}
