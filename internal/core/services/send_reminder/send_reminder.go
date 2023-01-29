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
			logging.Entry("status", prepared.Reminder.Status.String()),
		)
		return prepared, nil
	}

	err = s.sender.SendReminder(ctx, prepared.Reminder)
	status := reminder.StatusSentSuccess
	if err != nil {
		logging.Error(ctx, s.log, err, logging.Entry("input", input), logging.Entry("reminder", prepared.Reminder))
		status = reminder.StatusSentError
	}

	updatedReminder, err := s.reminderRepository.Update(
		ctx,
		reminder.UpdateInput{
			ID:             prepared.Reminder.ID,
			DoSentAtUpdate: true,
			SentAt:         c.NewOptional(s.now(), true),
			DoStatusUpdate: true,
			Status:         status,
		},
	)
	if err != nil {
		logging.Error(ctx, s.log, err, logging.Entry("input", input), logging.Entry("reminder", prepared.Reminder))
		return prepared, err
	}

	result.Reminder.FromReminderAndChannels(updatedReminder, prepared.Reminder.ChannelIDs)
	return result, err
}
