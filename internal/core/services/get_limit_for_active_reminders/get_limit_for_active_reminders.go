package getlimitforactivereminders

import (
	"context"
	c "remindme/internal/core/domain/common"
	e "remindme/internal/core/domain/errors"
	"remindme/internal/core/domain/logging"
	"remindme/internal/core/domain/reminder"
	"remindme/internal/core/domain/user"
	"remindme/internal/core/services"
	"remindme/internal/core/services/auth"
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
}

func New(
	log logging.Logger,
	limits user.LimitsRepository,
	reminders reminder.ReminderRepository,
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
	return &service{
		log:       log,
		limits:    limits,
		reminders: reminders,
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

	if !limits.ActiveReminderCount.IsPresent {
		return result, nil
	}

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

	result.Limit.IsPresent = true
	result.Limit.Value.Value = limits.ActiveReminderCount.Value
	result.Limit.Value.Actual = uint32(activeReminderCount)
	s.log.Info(
		ctx,
		"Got user limits for active reminder count.",
		logging.Entry("userID", input.UserID),
		logging.Entry("result", result),
	)
	return result, err
}
