package listuserreminders

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

const DEFAULT_LIMIT = 100

type Input struct {
	UserID   user.ID
	StatusIn c.Optional[[]reminder.Status]
	OrderBy  reminder.OrderBy
	Limit    c.Optional[uint]
	Offset   uint
}

func (i Input) WithAuthenticatedUser(u user.User) auth.Input {
	i.UserID = u.ID
	return i
}

type Result struct {
	Reminders  []reminder.ReminderWithChannels
	TotalCount uint
}

type service struct {
	log                logging.Logger
	reminderRepository reminder.ReminderRepository
}

func New(
	log logging.Logger,
	reminderRepository reminder.ReminderRepository,
) services.Service[Input, Result] {
	if log == nil {
		panic(e.NewNilArgumentError("log"))
	}
	if reminderRepository == nil {
		panic(e.NewNilArgumentError("channelRepository"))
	}
	return &service{
		log:                log,
		reminderRepository: reminderRepository,
	}
}

func (s *service) Run(ctx context.Context, input Input) (result Result, err error) {
	limit := c.NewOptional[uint](DEFAULT_LIMIT, true)
	if input.Limit.IsPresent {
		limit.Value = input.Limit.Value
	}

	readOptions := reminder.ReadOptions{
		CreatedByEquals: c.NewOptional(input.UserID, true),
		StatusIn:        input.StatusIn,
		Limit:           limit,
		Offset:          input.Offset,
		OrderBy:         input.OrderBy,
	}
	reminders, err := s.reminderRepository.Read(ctx, readOptions)
	if err != nil {
		logging.Error(ctx, s.log, err, logging.Entry("input", input))
		return result, err
	}
	totalCount, err := s.reminderRepository.Count(ctx, readOptions)
	if err != nil {
		logging.Error(ctx, s.log, err, logging.Entry("input", input))
		return result, err
	}

	s.log.Info(
		ctx,
		"User reminders successsfully read.",
		logging.Entry("input", input),
		logging.Entry("count", len(reminders)),
		logging.Entry("totalCount", totalCount),
	)
	result.Reminders = reminders
	result.TotalCount = totalCount
	return result, nil
}
