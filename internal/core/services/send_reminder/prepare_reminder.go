package sendreminder

import (
	"context"
	"errors"
	"math"
	c "remindme/internal/core/domain/common"
	e "remindme/internal/core/domain/errors"
	"remindme/internal/core/domain/logging"
	"remindme/internal/core/domain/reminder"
	uow "remindme/internal/core/domain/unit_of_work"
	"remindme/internal/core/domain/user"
	"remindme/internal/core/services"
	"time"
)

type Input struct {
	ReminderID reminder.ID
	At         time.Time
}

type Result struct {
	Reminder reminder.ReminderWithChannels
}

type prepareService struct {
	log        logging.Logger
	unitOfWork uow.UnitOfWork
	now        func() time.Time
}

func NewPrepareService(
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
	return &prepareService{
		log:        log,
		unitOfWork: unitOfWork,
		now:        now,
	}
}

func (s *prepareService) Run(ctx context.Context, input Input) (result Result, err error) {
	uow, err := s.unitOfWork.Begin(ctx)
	if err != nil {
		logging.Error(ctx, s.log, err, logging.Entry("input", input))
		return result, err
	}
	defer uow.Rollback(ctx)

	reminderRepo := uow.Reminders()
	err = reminderRepo.Lock(ctx, input.ReminderID)
	if err != nil {
		logging.Error(ctx, s.log, err, logging.Entry("input", input))
		return result, err
	}
	rem, err := reminderRepo.GetByID(ctx, input.ReminderID)
	if err != nil {
		logging.Error(ctx, s.log, err, logging.Entry("input", input))
		return result, err
	}
	if rem.Status != reminder.StatusScheduled {
		s.log.Info(
			ctx,
			"Reminder status is not Scheduled, skip sending.",
			logging.Entry("input", input),
			logging.Entry("status", rem.Status),
		)
		result.Reminder = rem
		return result, nil
	}
	if math.Abs(float64(rem.At.Sub(input.At.Truncate(time.Second)))) > float64(time.Second) {
		s.log.Info(
			ctx,
			"Reminder At time changed, skip sending.",
			logging.Entry("input", input),
			logging.Entry("at", rem.At),
		)
		result.Reminder = rem
		return result, nil
	}

	update := reminder.UpdateInput{
		ID:             rem.ID,
		DoStatusUpdate: true,
		Status:         reminder.StatusSending,
	}

	now := s.now()
	err = checkUserLimits(ctx, s.log, uow, rem.CreatedBy, now)
	if err != nil {
		switch {
		case errors.Is(err, user.ErrLimitSentReminderCountExceeded):
			update.Status = reminder.StatusSentLimitExceeded
			update.DoCanceledAtUpdate = true
			update.CanceledAt = c.NewOptional(now, true)
		default:
			return result, err
		}
	}

	if now.Sub(rem.At) > reminder.MAX_SENDING_DELAY {
		s.log.Error(
			ctx,
			"Sending delay exceeded, skip sending.",
			logging.Entry("input", input),
			logging.Entry("at", rem.At),
		)
		update.Status = reminder.StatusCanceled
		update.DoCanceledAtUpdate = true
		update.CanceledAt = c.NewOptional(now, true)
	}

	updatedReminder, err := reminderRepo.Update(ctx, update)
	if err != nil {
		logging.Error(ctx, s.log, err, logging.Entry("input", input))
		return result, err
	}

	if err := uow.Commit(ctx); err != nil {
		logging.Error(ctx, s.log, err, logging.Entry("input", input))
		return result, err
	}

	result.Reminder.FromReminderAndChannels(updatedReminder, rem.ChannelIDs)
	s.log.Info(
		ctx,
		"Reminder status has been successfully changed to 'sending'.",
		logging.Entry("input", input),
		logging.Entry("result", result),
	)
	return result, nil
}

func checkUserLimits(
	ctx context.Context,
	log logging.Logger,
	uow uow.Context,
	userID user.ID,
	now time.Time,
) error {
	limits, err := uow.Limits().GetUserLimits(ctx, userID)
	if err != nil {
		logging.Error(ctx, log, err, logging.Entry("userID", userID))
		return err
	}

	if !limits.MonthlySentReminderCount.IsPresent {
		log.Info(ctx, "User has no sent reminder count limit.", logging.Entry("userID", userID))
		return nil
	}

	sentAfter := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	logEntries := []logging.LogEntry{
		logging.Entry("userID", userID),
		logging.Entry("sendAfter", sentAfter),
		logging.Entry("limit", limits.MonthlySentReminderCount),
	}

	count, err := uow.Reminders().Count(ctx, reminder.ReadOptions{
		CreatedByEquals: c.NewOptional(userID, true),
		SentAfter:       c.NewOptional(sentAfter, true),
		StatusIn:        c.NewOptional([]reminder.Status{reminder.StatusSentSuccess}, true),
	})
	if err != nil {
		logging.Error(ctx, log, err, logEntries...)
		return err
	}

	logEntries = append(logEntries, logging.Entry("count", count))
	if count >= uint(limits.MonthlySentReminderCount.Value) {
		log.Info(ctx, "Sent reminder count limit exceeded.", logEntries...)
		return user.ErrLimitSentReminderCountExceeded
	}

	log.Info(ctx, "Reminder is allowed to be sent.", logEntries...)
	return nil
}
