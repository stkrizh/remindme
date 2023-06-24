package remindersender

import (
	"context"
	"encoding/json"
	"fmt"
	"remindme/internal/core/domain/channel"
	c "remindme/internal/core/domain/common"
	e "remindme/internal/core/domain/errors"
	"remindme/internal/core/domain/logging"
	"remindme/internal/core/domain/reminder"

	"github.com/r3labs/sse/v2"
)

type Sender struct {
	log            logging.Logger
	channelRepo    channel.Repository
	sseServer      *sse.Server
	emailSender    reminder.EmailSender
	telegramSender reminder.TelegramSender
	internalSender reminder.InternalSender
}

func New(
	log logging.Logger,
	channelRepo channel.Repository,
	sseServer *sse.Server,
	emailSender reminder.EmailSender,
	telegramSender reminder.TelegramSender,
	internalSender reminder.InternalSender,
) *Sender {
	if log == nil {
		panic(e.NewNilArgumentError("log"))
	}
	if channelRepo == nil {
		panic(e.NewNilArgumentError("channelRepo"))
	}
	if sseServer == nil {
		panic(e.NewNilArgumentError("sseServer"))
	}
	if emailSender == nil {
		panic(e.NewNilArgumentError("emailSender"))
	}
	if telegramSender == nil {
		panic(e.NewNilArgumentError("telegramSender"))
	}
	if internalSender == nil {
		panic(e.NewNilArgumentError("internalSender"))
	}
	return &Sender{
		log:            log,
		channelRepo:    channelRepo,
		sseServer:      sseServer,
		emailSender:    emailSender,
		telegramSender: telegramSender,
		internalSender: internalSender,
	}
}

func (s *Sender) SendReminder(ctx context.Context, rem reminder.ReminderWithChannels) error {
	channels, err := s.channelRepo.Read(ctx, channel.ReadOptions{IDIn: c.NewOptional(rem.ChannelIDs, true)})
	if err != nil {
		logging.Error(ctx, s.log, err, logging.Entry("reminder", rem))
		return err
	}

	s.log.Info(ctx, "Got channels for sending reminder.", logging.Entry("channels", channels))
	isInternalChannel := false
	for _, c := range channels {
		if c.Type == channel.Internal {
			isInternalChannel = true
		}
		channelSender := reminder.NewChannelSender(
			ctx,
			rem.Reminder,
			s.emailSender,
			s.telegramSender,
			s.internalSender,
		)
		err := channelSender.SendReminder(c.Settings)
		if err != nil {
			s.log.Error(
				ctx,
				"Could not send reminder.",
				logging.Entry("err", err),
				logging.Entry("reminder", rem),
				logging.Entry("channelID", c.ID),
				logging.Entry("channeSettings", c.Settings),
			)
		} else {
			s.log.Info(
				ctx,
				"Reminder has been successfully sent to channel.",
				logging.Entry("reminderID", rem.ID),
				logging.Entry("channelID", c.ID),
				logging.Entry("channeSettings", c.Settings),
			)
		}
	}
	s.publishSse(rem, isInternalChannel)

	if err != nil {
		s.log.Error(
			ctx,
			"Could not send reminder.",
			logging.Entry("err", err),
			logging.Entry("reminder", rem),
			logging.Entry("channels", channels),
		)
		return err
	}
	s.log.Info(ctx, "Reminder has been sent.", logging.Entry("reminderID", rem.ID))
	return nil
}

func (s *Sender) publishSse(rem reminder.ReminderWithChannels, isInternalChannel bool) {
	var event []byte
	if isInternalChannel {
		event = []byte("internalReminderSent")
	} else {
		event = []byte("reminderSent")
	}

	sseData, _ := json.Marshal(
		ReminderEvent{
			ID:   int64(rem.ID),
			Body: rem.Body,
		},
	)
	s.sseServer.Publish(fmt.Sprintf("%d", rem.CreatedBy), &sse.Event{
		Event: event,
		Data:  sseData,
	})
}

type ReminderEvent struct {
	ID   int64  `json:"id"`
	Body string `json:"body"`
}
