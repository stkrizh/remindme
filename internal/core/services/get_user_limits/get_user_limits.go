package getuserlimits

import (
	"context"
	"remindme/internal/core/domain/channel"
	c "remindme/internal/core/domain/common"
	e "remindme/internal/core/domain/errors"
	"remindme/internal/core/domain/logging"
	"remindme/internal/core/domain/reminder"
	"remindme/internal/core/domain/user"
	"remindme/internal/core/services"
	"remindme/internal/core/services/auth"
	"time"
)

type Input struct {
	UserID user.ID
}

func (i Input) WithAuthenticatedUser(u user.User) auth.Input {
	i.UserID = u.ID
	return i
}

type Result struct {
	Limits user.Limits
	Values user.Limits
}

type service struct {
	log       logging.Logger
	limits    user.LimitsRepository
	channels  channel.Repository
	reminders reminder.ReminderRepository
	now       func() time.Time
}

func New(
	log logging.Logger,
	limits user.LimitsRepository,
	channels channel.Repository,
	reminders reminder.ReminderRepository,
	now func() time.Time,
) services.Service[Input, Result] {
	if log == nil {
		panic(e.NewNilArgumentError("log"))
	}
	if limits == nil {
		panic(e.NewNilArgumentError("limits"))
	}
	if channels == nil {
		panic(e.NewNilArgumentError("channels"))
	}
	if reminders == nil {
		panic(e.NewNilArgumentError("reminders"))
	}
	if now == nil {
		panic(e.NewNilArgumentError("now"))
	}
	return &service{
		log:       log,
		limits:    limits,
		channels:  channels,
		reminders: reminders,
		now:       now,
	}
}

func (s *service) Run(ctx context.Context, input Input) (result Result, err error) {
	defer func() {
		if err != nil {
			logging.Error(ctx, s.log, err, logging.Entry("input", input))
		}
	}()

	limits, err := s.limits.GetUserLimits(ctx, input.UserID)
	if err != nil {
		return result, err
	}

	if limits.EmailChannelCount.IsPresent || limits.TelegramChannelCount.IsPresent {
		channels, err := s.channels.Read(ctx, channel.ReadOptions{UserIDEquals: c.NewOptional(input.UserID, true)})
		if err != nil {
			return result, err
		}
		s.countChannels(limits, channels, &result)
	}

	if limits.ActiveReminderCount.IsPresent {
		activeReminderCount, err := s.reminders.Count(
			ctx,
			reminder.ReadOptions{
				CreatedByEquals: c.NewOptional(input.UserID, true),
				StatusIn: c.NewOptional(
					[]reminder.Status{reminder.StatusCreated, reminder.StatusScheduled},
					true,
				),
			},
		)
		if err != nil {
			return result, err
		}
		result.Values.ActiveReminderCount = c.NewOptional(uint32(activeReminderCount), true)
	}

	if limits.MonthlySentReminderCount.IsPresent {
		now := s.now()
		sentAfter := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)

		sentReminderCount, err := s.reminders.Count(
			ctx,
			reminder.ReadOptions{
				CreatedByEquals: c.NewOptional(input.UserID, true),
				StatusIn:        c.NewOptional([]reminder.Status{reminder.StatusSentSuccess}, true),
				SentAfter:       c.NewOptional(sentAfter, true),
			},
		)
		if err != nil {
			return result, err
		}
		result.Values.MonthlySentReminderCount = c.NewOptional(uint32(sentReminderCount), true)
	}

	s.log.Info(
		ctx,
		"User limits successfully calculated.",
		logging.Entry("userID", input.UserID),
		logging.Entry("result", result),
	)
	result.Limits = limits
	return result, err
}

func (s *service) countChannels(limits user.Limits, channels []channel.Channel, result *Result) {
	if limits.EmailChannelCount.IsPresent {
		result.Values.EmailChannelCount = c.NewOptional(uint32(0), true)
	}
	if limits.TelegramChannelCount.IsPresent {
		result.Values.TelegramChannelCount = c.NewOptional(uint32(0), true)
	}
	for _, ch := range channels {
		switch {
		case ch.Type == channel.Email && limits.EmailChannelCount.IsPresent:
			result.Values.EmailChannelCount.Value++
		case ch.Type == channel.Telegram && limits.TelegramChannelCount.IsPresent:
			result.Values.TelegramChannelCount.Value++
		}
	}
}
