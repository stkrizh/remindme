package deletereminder

import (
	"context"
	"errors"
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
	UserID     user.ID
	ReminderID reminder.ID
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
	now        func() time.Time
}

func New(
	log logging.Logger,
	unitOfWork uow.UnitOfWork,
	now func() time.Time,
) services.Service[Input, Result] {
	if log == nil {
		panic(e.NewNilArgumentError("log"))
	}
	if unitOfWork == nil {
		panic(e.NewNilArgumentError("unitOfWork"))
	}
	if now == nil {
		panic(e.NewNilArgumentError("now"))
	}
	return &service{
		log:        log,
		unitOfWork: unitOfWork,
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
		s.log.Info(ctx, "Reminder is not active and can't be deleted.", logging.Entry("input", input))
		return result, reminder.ErrReminderNotActive
	}

	err = reminderRepository.Delete(ctx, rem.ID)
	if err != nil {
		switch {
		case errors.Is(err, reminder.ErrReminderDoesNotExist):
			// do nothing
		default:
			logging.Error(ctx, s.log, err, logging.Entry("input", input))
		}
		return result, err
	}

	if err := uow.Commit(ctx); err != nil {
		logging.Error(ctx, s.log, err, logging.Entry("input", input))
		return result, err
	}

	s.log.Info(
		ctx,
		"Reminder has been successfully deleted.",
		logging.Entry("input", input),
		logging.Entry("reminderID", rem.ID),
	)
	result.Reminder = rem
	return result, nil
}
