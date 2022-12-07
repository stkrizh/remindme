package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"net/netip"
	ratelimiter "remindme/internal/domain/rate_limiter"
	"remindme/internal/domain/services"
	signupanonymously "remindme/internal/domain/services/sign_up_anonymously"
)

const IP_HEADER = "X-Real-IP"

type SignUpAnonymously struct {
	service services.Service[signupanonymously.Input, signupanonymously.Result]
}

func NewSignUpAnonymously(
	service services.Service[signupanonymously.Input, signupanonymously.Result],
) *SignUpAnonymously {
	return &SignUpAnonymously{service: service}
}

type SignUpAnonymouslyResult struct {
	Identity string `json:"identity"`
	Token    string `json:"token"`
}

func (s *SignUpAnonymously) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	ipRaw := r.Header.Get(IP_HEADER)
	if ipRaw == "" {
		renderErrorResponse(rw, fmt.Sprintf("%s header must be provided", IP_HEADER), http.StatusBadRequest)
		return
	}
	ip, err := netip.ParseAddr(ipRaw)
	if err != nil {
		renderErrorResponse(rw, fmt.Sprintf("invalid %s header value", IP_HEADER), http.StatusBadRequest)
		return
	}

	result, err := s.service.Run(r.Context(), signupanonymously.Input{IP: ip})
	if err == nil {
		renderResponse(
			rw,
			SignUpAnonymouslyResult{
				Identity: string(result.User.Identity.Value),
				Token:    string(result.Token),
			},
			http.StatusCreated,
		)
		return
	}

	if errors.Is(err, ratelimiter.ErrRateLimitExceeded) {
		renderErrorResponse(rw, "rate limit exceeded", http.StatusTooManyRequests)
		return
	}

	renderErrorResponse(rw, "internal error", http.StatusInternalServerError)
}
