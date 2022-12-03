package handlers

import (
	"encoding/json"
	"net/http"
)

type errorResponse struct {
	Error string `json:"error"`
}

func renderErrorResponse(rw http.ResponseWriter, msg string, status int) {
	renderResponse(rw, errorResponse{Error: msg}, status)
}

func renderResponse(rw http.ResponseWriter, res interface{}, status int) {
	rw.Header().Set("Content-Type", "application/json")

	content, err := json.Marshal(res)
	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	rw.WriteHeader(status)
	rw.Write(content)
}
