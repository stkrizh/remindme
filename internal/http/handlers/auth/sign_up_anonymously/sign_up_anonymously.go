package handlers

import (
	"fmt"
	"net/http"
	"net/netip"
	"remindme/internal/core/services"
	signupanonymously "remindme/internal/core/services/sign_up_anonymously"
	"remindme/internal/http/handlers/response"
)

const IP_HEADER = "X-Real-IP"

type Handler struct {
	service services.Service[signupanonymously.Input, signupanonymously.Result]
}

func New(
	service services.Service[signupanonymously.Input, signupanonymously.Result],
) *Handler {
	return &Handler{service: service}
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

	result, err := h.service.Run(r.Context(), signupanonymously.Input{IP: ip})
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
