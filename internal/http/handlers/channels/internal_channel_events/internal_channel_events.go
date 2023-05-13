package createemailchannel

import (
	"net/http"
	"remindme/internal/core/domain/channel"
	e "remindme/internal/core/domain/errors"
	"remindme/internal/core/domain/logging"
	"remindme/internal/http/handlers/response"

	"github.com/r3labs/sse/v2"
)

type Handler struct {
	log            logging.Logger
	sseServer      *sse.Server
	tokenValidator channel.InternalChannelTokenValidator
}

func New(
	log logging.Logger,
	sseServer *sse.Server,
	tokenValidator channel.InternalChannelTokenValidator,
) *Handler {
	if log == nil {
		panic(e.NewNilArgumentError("log"))
	}
	if sseServer == nil {
		panic(e.NewNilArgumentError("sseServer"))
	}
	if tokenValidator == nil {
		panic(e.NewNilArgumentError("tokenValidator"))
	}
	return &Handler{log: log, sseServer: sseServer, tokenValidator: tokenValidator}
}

func (h *Handler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	streamID := r.URL.Query().Get("stream")
	internalChannelToken := channel.InternalChannelToken(streamID)

	if !h.tokenValidator.ValidateInternalChannelToken(internalChannelToken) {
		h.log.Info(r.Context(), "Invalid internal channel token.", logging.Entry("token", internalChannelToken))
		response.RenderError(rw, "invalid internal channel token", http.StatusNotFound)
		return
	}

	go func() {
		// Received browser disconnection
		<-r.Context().Done()
		h.log.Info(
			r.Context(),
			"Unsubscribed from internal channel events.",
			logging.Entry("token", internalChannelToken),
		)
		h.sseServer.RemoveStream(streamID)
	}()

	h.sseServer.CreateStream(streamID)
	h.log.Info(
		r.Context(),
		"Subscribed to internal channel events.",
		logging.Entry("token", internalChannelToken),
	)
	h.sseServer.ServeHTTP(rw, r)
}
