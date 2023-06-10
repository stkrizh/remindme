package sendreminder

import (
	"context"
	c "remindme/internal/core/domain/common"
	e "remindme/internal/core/domain/errors"
	"remindme/internal/core/domain/logging"
	"remindme/internal/core/domain/reminder"
	"remindme/internal/core/services"
	"time"
)

type sendService struct {
	log                logging.Logger
	reminderRepository reminder.ReminderRepository
	sender             reminder.Sender
	now                func() time.Time
	prepareService     services.Service[Input, Result]
}

func NewSendService(
	log logging.Logger,
	reminderRepository reminder.ReminderRepository,
	sender reminder.Sender,
	now func() time.Time,
	prepareService services.Service[Input, Result],
) services.Service[Input, Result] {
	if log == nil {
		panic(e.NewNilArgumentError("log"))
	}
	if reminderRepository == nil {
		panic(e.NewNilArgumentError("reminderRepository"))
	}
	if sender == nil {
		panic(e.NewNilArgumentError("sender"))
	}
	if now == nil {
		panic(e.NewNilArgumentError("now"))
	}
	if prepareService == nil {
		panic(e.NewNilArgumentError("prepareService"))
	}
	return &sendService{
		log:                log,
		reminderRepository: reminderRepository,
		sender:             sender,
		now:                now,
		prepareService:     prepareService,
	}
}

func (s *sendService) Run(ctx context.Context, input Input) (result Result, err error) {
	prepared, err := s.prepareService.Run(ctx, input)
	if err != nil {
		return prepared, err
	}

	if prepared.Reminder.Status != reminder.StatusSending {
		s.log.Info(
			ctx,
			"Reminder is skipped due to the status is not 'sending'.",
			logging.Entry("input", input),
			logging.Entry("status", prepared.Reminder.Status),
		)
		return prepared, nil
	}

	update := reminder.UpdateInput{
		ID:             prepared.Reminder.ID,
		DoStatusUpdate: true,
		Status:         reminder.StatusSentSuccess,
	}
	if s.now().Sub(prepared.Reminder.At) > reminder.MAX_SENDING_DELAY {
		s.log.Error(
			ctx,
			"Sending delay exceeded, skip sending.",
			logging.Entry("input", input),
			logging.Entry("at", prepared.Reminder.At),
		)
		update.Status = reminder.StatusCanceled
		update.DoCanceledAtUpdate = true
		update.CanceledAt = c.NewOptional(s.now(), true)
	}

	if update.Status != reminder.StatusCanceled {
		err = s.sender.SendReminder(ctx, prepared.Reminder)
		update.DoSentAtUpdate = true
		update.SentAt = c.NewOptional(s.now(), true)
		if err != nil {
			logging.Error(ctx, s.log, err, logging.Entry("input", input), logging.Entry("reminder", prepared.Reminder))
			update.Status = reminder.StatusSentError
		}
	}

	updatedReminder, err := s.reminderRepository.Update(ctx, update)
	if err != nil {
		logging.Error(ctx, s.log, err, logging.Entry("input", input), logging.Entry("reminder", prepared.Reminder))
		return prepared, err
	}

	result.Reminder.FromReminderAndChannels(updatedReminder, prepared.Reminder.ChannelIDs)
	return result, err
}
