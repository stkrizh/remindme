package remindersender

import (
	"context"
	"remindme/internal/core/domain/channel"
	c "remindme/internal/core/domain/common"
	e "remindme/internal/core/domain/errors"
	"remindme/internal/core/domain/logging"
	"remindme/internal/core/domain/reminder"
)

type Sender struct {
	log            logging.Logger
	channelRepo    channel.Repository
	emailSender    reminder.EmailSender
	telegramSender reminder.TelegramSender
	internalSender reminder.InternalSender
}

func New(
	log logging.Logger,
	channelRepo channel.Repository,
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
	for _, channel := range channels {
		channelSender := reminder.NewChannelSender(
			ctx,
			rem.Reminder,
			s.emailSender,
			s.telegramSender,
			s.internalSender,
		)
		err := channelSender.SendReminder(channel.Settings)
		if err != nil {
			logging.Error(
				ctx,
				s.log,
				err,
				logging.Entry("reminder", rem),
				logging.Entry("channeID", channel.ID),
				logging.Entry("channeSettings", channel.Settings),
			)
		} else {
			s.log.Info(
				ctx,
				"Reminder has been successfully sent to channel.",
				logging.Entry("reminderID", rem.ID),
				logging.Entry("channelID", channel.ID),
				logging.Entry("channeSettings", channel.Settings),
			)
		}
	}
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
