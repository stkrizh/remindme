package response

import (
	"encoding/json"
	"net/http"
)

type errorResponse struct {
	Error string `json:"error"`
}

func RenderUnauthorized(rw http.ResponseWriter) {
	RenderError(rw, "invalid authentication token", http.StatusUnauthorized)
}

func RenderInternalError(rw http.ResponseWriter) {
	RenderError(rw, "internal error", http.StatusInternalServerError)
}

func RenderRateLimitExceeded(rw http.ResponseWriter) {
	RenderError(rw, "rate limit exceeded", http.StatusTooManyRequests)
}

func RenderError(rw http.ResponseWriter, msg string, status int) {
	Render(rw, errorResponse{Error: msg}, status)
}

func Render(rw http.ResponseWriter, res interface{}, status int) {
	rw.Header().Set("Content-Type", "application/json")

	content, err := json.Marshal(res)
	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	rw.WriteHeader(status)
	rw.Write(content)
}
