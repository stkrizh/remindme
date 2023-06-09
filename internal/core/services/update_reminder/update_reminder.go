package updatereminder

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
	"remindme/internal/core/services/auth"
	"time"
)

type Input struct {
	UserID        user.ID
	ReminderID    reminder.ID
	DoAtUpdate    bool
	At            time.Time
	DoEveryUpdate bool
	Every         c.Optional[reminder.Every]
	DoBodyUpdate  bool
	Body          string
}

func (i Input) WithAuthenticatedUser(u user.User) auth.Input {
	i.UserID = u.ID
	return i
}

type Result struct {
	Reminder reminder.ReminderWithChannels
}

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

	reminderRepository := uow.Reminders()
	reminderRepository.Lock(ctx, input.ReminderID)
	rem, err := reminderRepository.GetByID(ctx, input.ReminderID)
	if err != nil {
		switch {
		case errors.Is(err, reminder.ErrReminderDoesNotExist):
			s.log.Info(ctx, "Reminder not found.", logging.Entry("input", input))
		default:
			logging.Error(ctx, s.log, err, logging.Entry("input", input))
		}
		return result, err
	}

	if rem.CreatedBy != input.UserID {
		s.log.Info(ctx, "Reminder belongs to another user.", logging.Entry("input", input))
		return result, reminder.ErrReminderPermission
	}
	if !rem.IsActive() {
		s.log.Info(ctx, "Reminder is not active and can't be updated.", logging.Entry("input", input))
		return result, reminder.ErrReminderNotActive
	}

	now := s.now()
	doStatusUpdate := false
	status := rem.Status
	doScheduledAtUpdate := false
	scheduledAt := rem.ScheduledAt
	if doAtUpdate(input, rem.At) {
		if err := validateAt(input.At, now); err != nil {
			return result, err
		}
		doStatusUpdate = true
		doScheduledAtUpdate = true
		if input.At.Sub(now) < reminder.DURATION_FOR_SCHEDULING {
			status = reminder.StatusScheduled
			scheduledAt = c.NewOptional(now, true)
		} else {
			status = reminder.StatusCreated
			scheduledAt = c.Optional[time.Time]{}
		}
	}

	if input.DoEveryUpdate {
		if err := validateEvery(ctx, s.log, input.UserID, input.Every, uow.Limits()); err != nil {
			return result, err
		}
	}

	updatedReminder, err := uow.Reminders().Update(
		ctx,
		reminder.UpdateInput{
			ID:                  input.ReminderID,
			DoAtUpdate:          input.DoAtUpdate,
			At:                  input.At,
			DoEveryUpdate:       input.DoEveryUpdate,
			Every:               input.Every,
			DoBodyUpdate:        input.DoBodyUpdate,
			Body:                input.Body,
			DoStatusUpdate:      doStatusUpdate,
			Status:              status,
			DoScheduledAtUpdate: doScheduledAtUpdate,
			ScheduledAt:         scheduledAt,
		},
	)
	if err != nil {
		logging.Error(ctx, s.log, err, logging.Entry("input", input))
		return result, err
	}

	if doStatusUpdate && updatedReminder.Status == reminder.StatusScheduled {
		if err := s.scheduler.ScheduleReminder(ctx, updatedReminder); err != nil {
			logging.Error(ctx, s.log, err, logging.Entry("input", input))
			return result, err
		}
	}

	if err := uow.Commit(ctx); err != nil {
		logging.Error(ctx, s.log, err, logging.Entry("input", input))
		return result, err
	}

	s.log.Info(
		ctx,
		"Reminder successfully updated.",
		logging.Entry("input", input),
		logging.Entry("reminder", updatedReminder),
	)

	result.Reminder.FromReminderAndChannels(updatedReminder, rem.ChannelIDs)
	return result, nil
}

func validateAt(at time.Time, now time.Time) error {
	if at.Location() != time.UTC {
		return reminder.ErrReminderAtTimeIsNotUTC
	}
	duration_from_now := at.Sub(now)
	if duration_from_now < reminder.MIN_DURATION_FROM_NOW {
		return reminder.ErrReminderTooEarly
	}
	if duration_from_now > reminder.MAX_DURATION_FROM_NOW {
		return reminder.ErrReminderTooLate
	}
	return nil
}

func validateEvery(
	ctx context.Context,
	log logging.Logger,
	userID user.ID,
	every c.Optional[reminder.Every],
	limitsRepository user.LimitsRepository,
) error {
	if !every.IsPresent {
		return nil
	}
	if err := every.Value.Validate(); err != nil {
		return err
	}
	userLimits, err := limitsRepository.GetUserLimits(ctx, userID)
	if err != nil {
		logging.Error(ctx, log, err, logging.Entry("userID", userID))
		return err
	}
	if userLimits.ReminderEveryPerDayCount.IsPresent &&
		every.Value.PerDayCount() > userLimits.ReminderEveryPerDayCount.Value {
		log.Info(
			ctx,
			"Could not update reminder, every per day count limit exceeded.",
			logging.Entry("userID", userID),
			logging.Entry("every", every),
		)
		return user.ErrLimitReminderEveryPerDayCountExceeded
	}

	return nil
}

func doAtUpdate(input Input, currentAt time.Time) bool {
	return input.DoAtUpdate && input.At.Round(time.Second) != currentAt.Round(time.Second)
}
