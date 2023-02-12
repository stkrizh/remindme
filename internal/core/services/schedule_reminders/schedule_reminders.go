package schedulereminders

import (
	"context"
	e "remindme/internal/core/domain/errors"
	"remindme/internal/core/domain/logging"
	"remindme/internal/core/domain/reminder"
	uow "remindme/internal/core/domain/unit_of_work"
	"remindme/internal/core/services"
	"time"
)

type Input struct{}

type Result struct{}

type service struct {
	log        logging.Logger
	unitOfWork uow.UnitOfWork
	scheduler  reminder.Scheduler
	now        func() time.Time
}

func New(
	log logging.Logger,
	unitOfWork uow.UnitOfWork,
	scheduler reminder.Scheduler,
	now func() time.Time,
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
	if now == nil {
		panic(e.NewNilArgumentError("now"))
	}
	return &service{
		log:        log,
		unitOfWork: unitOfWork,
		scheduler:  scheduler,
		now:        now,
	}
}

func (s *service) Run(ctx context.Context, input Input) (result Result, err error) {
	uow, err := s.unitOfWork.Begin(ctx)
	if err != nil {
		logging.Error(ctx, s.log, err, logging.Entry("input", input))
		return result, err
	}
	defer uow.Rollback(ctx)

	now := s.now()
	scheduledReminders, err := uow.Reminders().Schedule(
		ctx,
		reminder.ScheduleInput{
			AtBefore:    now.Add(reminder.DURATION_FOR_SCHEDULING),
			ScheduledAt: now,
		},
	)
	if err != nil {
		logging.Error(ctx, s.log, err, logging.Entry("input", input))
		return result, err
	}

	s.log.Info(
		ctx,
		"Got reminders for scheduling.",
		logging.Entry("count", len(scheduledReminders)),
	)
	scheduledIDs := make([]reminder.ID, 0, len(scheduledReminders))
	for ix, reminder := range scheduledReminders {
		err := s.scheduler.ScheduleReminder(ctx, reminder)
		if err != nil {
			logging.Error(
				ctx,
				s.log,
				err,
				logging.Entry("index", ix),
				logging.Entry("reminderID", reminder.ID),
				logging.Entry("scheduledIDs", scheduledIDs),
			)
			return result, err
		}
		scheduledIDs = append(scheduledIDs, reminder.ID)
	}

	if err := uow.Commit(ctx); err != nil {
		logging.Error(ctx, s.log, err)
		return result, err
	}

	if len(scheduledIDs) > 0 {
		s.log.Info(
			ctx,
			"Reminders successfully scheduled.",
			logging.Entry("scheduledCount", len(scheduledIDs)),
			logging.Entry("scheduledIDs", scheduledIDs),
		)
	}
	return result, nil
}
