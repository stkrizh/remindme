package sendreminder

import (
	"context"
	"errors"
	c "remindme/internal/core/domain/common"
	e "remindme/internal/core/domain/errors"
	"remindme/internal/core/domain/logging"
	"remindme/internal/core/domain/reminder"
	uow "remindme/internal/core/domain/unit_of_work"
	"remindme/internal/core/domain/user"
	"remindme/internal/core/services"
)

type createNextPeriodicService struct {
	log            logging.Logger
	unitOfWork     uow.UnitOfWork
	scheduler      reminder.Scheduler
	prepareService services.Service[Input, Result]
}

func NewCreateNextPeriodicService(
	log logging.Logger,
	unitOfWork uow.UnitOfWork,
	scheduler reminder.Scheduler,
	prepareService services.Service[Input, Result],
) services.Service[Input, Result] {
	if log == nil {
		panic(e.NewNilArgumentError("log"))
	}
	if unitOfWork == nil {
		panic(e.NewNilArgumentError("unitOfWork"))
	}
	if scheduler == nil {
		panic(e.NewNilArgumentError("scheduler"))
	}
	if prepareService == nil {
		panic(e.NewNilArgumentError("prepareService"))
	}

	return &createNextPeriodicService{
		log:            log,
		unitOfWork:     unitOfWork,
		scheduler:      scheduler,
		prepareService: prepareService,
	}
}

func (s *createNextPeriodicService) Run(ctx context.Context, input Input) (result Result, err error) {
	result, err = s.prepareService.Run(ctx, input)
	if !result.Reminder.Every.IsPresent {
		s.log.Info(
			ctx,
			"Reminder is not periodic, skip the next reminder creation.",
			logging.Entry("input", input),
			logging.Entry("every", result.Reminder.Every),
		)
		return result, err
	}

	if err != nil && !errors.Is(err, user.ErrLimitSentReminderCountExceeded) {
		s.log.Error(
			ctx,
			"Prepare service returned an error, could not create the next reminder.",
			logging.Entry("input", input),
			logging.Entry("err", err),
		)
		return result, err
	}

	uow, err := s.unitOfWork.Begin(ctx)
	if err != nil {
		logging.Error(ctx, s.log, err, logging.Entry("input", input))
		return result, err
	}
	defer uow.Rollback(ctx)

	status := reminder.StatusCreated
	scheduledAt := c.NewOptional(result.Reminder.At, false)
	if result.Reminder.Every.Value.TotalDuration() < reminder.DURATION_FOR_SCHEDULING {
		status = reminder.StatusScheduled
		scheduledAt.IsPresent = true
	}
	nextReminder, err := uow.Reminders().Create(ctx, reminder.CreateInput{
		CreatedBy:   result.Reminder.CreatedBy,
		CreatedAt:   result.Reminder.CreatedAt,
		At:          result.Reminder.Every.Value.NextFrom(result.Reminder.At),
		Body:        result.Reminder.Body,
		Every:       result.Reminder.Every,
		Status:      status,
		ScheduledAt: scheduledAt,
	})
	if err != nil {
		logging.Error(ctx, s.log, err, logging.Entry("result", result), logging.Entry("err", err))
		return result, err
	}

	nextReminderChannelIDs, err := uow.ReminderChannels().Create(
		ctx,
		reminder.CreateChannelsInput{
			ReminderID: nextReminder.ID,
			ChannelIDs: reminder.NewChannelIDs(result.Reminder.ChannelIDs...),
		},
	)
	if err != nil {
		logging.Error(ctx, s.log, err, logging.Entry("result", result), logging.Entry("err", err))
		return result, err
	}

	if err := uow.Commit(ctx); err != nil {
		logging.Error(ctx, s.log, err, logging.Entry("result", result), logging.Entry("err", err))
		return result, err
	}

	if nextReminder.Status == reminder.StatusScheduled {
		if err := s.scheduler.ScheduleReminder(ctx, nextReminder); err != nil {
			logging.Error(ctx, s.log, err, logging.Entry("result", result), logging.Entry("err", err))
		}
	}
	s.log.Info(
		ctx,
		"Next periodic reminder created.",
		logging.Entry("nextReminder", nextReminder),
		logging.Entry("nextReminderChannelIDs", nextReminderChannelIDs),
	)
	return result, nil
}
