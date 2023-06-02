package getlimitforsentreminders

import (
	"context"
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
	Limit c.Optional[user.Limit]
}

type service struct {
	log       logging.Logger
	limits    user.LimitsRepository
	reminders reminder.ReminderRepository
	now       func() time.Time
}

func New(
	log logging.Logger,
	limits user.LimitsRepository,
	reminders reminder.ReminderRepository,
	now func() time.Time,
) services.Service[Input, Result] {
	if log == nil {
		panic(e.NewNilArgumentError("log"))
	}
	if limits == nil {
		panic(e.NewNilArgumentError("limits"))
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

	if !limits.MonthlySentReminderCount.IsPresent {
		return result, nil
	}

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

	result.Limit.IsPresent = true
	result.Limit.Value.Value = uint32(limits.MonthlySentReminderCount.Value)
	result.Limit.Value.Actual = uint32(sentReminderCount)
	s.log.Info(
		ctx,
		"Got user limits for monthly sent reminders.",
		logging.Entry("userID", input.UserID),
		logging.Entry("result", result),
	)
	return result, err
}
