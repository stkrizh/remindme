package remindersender

import (
	"context"
	"fmt"
	"remindme/internal/core/domain/channel"
	e "remindme/internal/core/domain/errors"
	"remindme/internal/core/domain/reminder"

	"github.com/r3labs/sse/v2"
)

type InternalSender struct {
	sseServer *sse.Server
}

func NewInternal(sseServer *sse.Server) *InternalSender {
	if sseServer == nil {
		panic(e.NewNilArgumentError("sseServer"))
	}
	return &InternalSender{
		sseServer: sseServer,
	}
}

func (s *InternalSender) SendReminder(
	ctx context.Context,
	rem reminder.Reminder,
	settings *channel.InternalSettings,
) error {
	s.sseServer.Publish(string(settings.Token), &sse.Event{
		Event: []byte("reminder"),
		Data:  []byte(fmt.Sprintf("%d:%s", rem.ID, rem.Body)),
	})
	return nil
}
