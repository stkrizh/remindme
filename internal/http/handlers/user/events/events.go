package events

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	e "remindme/internal/core/domain/errors"
	"remindme/internal/core/domain/logging"
	"remindme/internal/core/domain/user"
	"remindme/internal/core/services"
	domainAuth "remindme/internal/core/services/auth"
	s "remindme/internal/core/services/get_user_by_session_token"
	"remindme/internal/http/handlers/auth"
	"remindme/internal/http/handlers/response"

	"github.com/go-chi/chi/v5"
	"github.com/r3labs/sse/v2"
)

type Handler struct {
	log       logging.Logger
	service   services.Service[s.Input, s.Result]
	sseServer *sse.Server
}

func New(
	log logging.Logger,
	sseServer *sse.Server,
	service services.Service[s.Input, s.Result],
) *Handler {
	if log == nil {
		panic(e.NewNilArgumentError("log"))
	}
	if sseServer == nil {
		panic(e.NewNilArgumentError("sseServer"))
	}
	if service == nil {
		panic(e.NewNilArgumentError("service"))
	}
	return &Handler{log: log, sseServer: sseServer, service: service}
}

func (h *Handler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	tokenFromURLParam := chi.URLParam(r, "sessionToken")
	if tokenFromURLParam != "" {
		if len(tokenFromURLParam) > auth.AUTH_TOKEN_MAX_LEN {
			response.RenderUnauthorized(rw)
			return
		}
		ctx := context.WithValue(r.Context(), domainAuth.CONTEXT_AUTH_TOKEN_KEY, user.SessionToken(tokenFromURLParam))
		r = r.WithContext(ctx)
	}

	result, err := h.service.Run(r.Context(), s.Input{})
	if err != nil {
		switch {
		case errors.Is(err, user.ErrUserDoesNotExist):
			response.RenderUnauthorized(rw)
		default:
			response.RenderInternalError(rw)
		}
		return
	}

	streamID := r.URL.Query().Get("stream")
	if streamID != fmt.Sprintf("%d", result.User.ID) {
		response.RenderError(rw, "invalid stream", http.StatusBadRequest)
		return
	}

	go func() {
		// Received browser disconnection
		<-r.Context().Done()
		h.log.Info(
			r.Context(),
			"Unsubscribed from user events.",
			logging.Entry("userID", result.User.ID),
		)
		h.sseServer.RemoveStream(streamID)
	}()

	h.log.Info(
		r.Context(),
		"Subscribed to user events.",
		logging.Entry("userID", result.User.ID),
		logging.Entry("streamID", streamID),
	)
	h.sseServer.ServeHTTP(rw, r)
}
