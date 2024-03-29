package createreminder

import (
	"context"
	"remindme/internal/core/domain/channel"
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
	UserID     user.ID
	At         time.Time
	Body       string
	Every      c.Optional[reminder.Every]
	ChannelIDs reminder.ChannelIDs
}

func (i Input) Validate(now time.Time) error {
	if i.At.Location() != time.UTC {
		return reminder.ErrReminderAtTimeIsNotUTC
	}
	duration_from_now := i.At.Sub(now)
	if duration_from_now < reminder.MIN_DURATION_FROM_NOW {
		return reminder.ErrReminderTooEarly
	}
	if duration_from_now > reminder.MAX_DURATION_FROM_NOW {
		return reminder.ErrReminderTooLate
	}
	if i.Every.IsPresent {
		if err := i.Every.Value.Validate(); err != nil {
			return err
		}
	}
	if len(i.ChannelIDs) == 0 {
		return reminder.ErrReminderChannelsNotSet
	}
	if len(i.ChannelIDs) > reminder.MAX_CHANNEL_COUNT {
		return reminder.ErrReminderTooManyChannels
	}
	return nil
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
	err = input.Validate(s.now())
	if err != nil {
		return result, err
	}

	uow, err := s.unitOfWork.Begin(ctx)
	if err != nil {
		logging.Error(ctx, s.log, err, logging.Entry("input", input))
		return result, err
	}
	defer uow.Rollback(ctx)

	userLimits, err := uow.Limits().GetUserLimitsWithLock(ctx, input.UserID)
	if err != nil {
		logging.Error(ctx, s.log, err, logging.Entry("input", input))
		return result, err
	}
	err = s.checkUserLimits(ctx, uow, userLimits, input)
	if err != nil {
		return result, err
	}

	channelIDs, err := s.readChannels(ctx, uow, input)
	if err != nil {
		return result, err
	}

	createInput := reminder.CreateInput{
		CreatedBy: input.UserID,
		CreatedAt: s.now(),
		Body:      input.Body,
		At:        input.At,
		Every:     input.Every,
		Status:    reminder.StatusCreated,
	}
	if input.At.Sub(s.now()) < reminder.DURATION_FOR_SCHEDULING {
		createInput.Status = reminder.StatusScheduled
		createInput.ScheduledAt = c.NewOptional(createInput.CreatedAt, true)
	}
	createdReminder, err := uow.Reminders().Create(ctx, createInput)
	if err != nil {
		logging.Error(ctx, s.log, err, logging.Entry("input", input))
		return result, err
	}
	_, err = uow.ReminderChannels().Create(ctx, reminder.CreateChannelsInput{
		ReminderID: createdReminder.ID,
		ChannelIDs: input.ChannelIDs,
	})
	if err != nil {
		logging.Error(ctx, s.log, err, logging.Entry("input", input), logging.Entry("reminder", createdReminder))
		return result, err
	}

	if createdReminder.Status == reminder.StatusScheduled {
		if err := s.scheduler.ScheduleReminder(ctx, createdReminder); err != nil {
			logging.Error(ctx, s.log, err, logging.Entry("input", input), logging.Entry("reminder", createdReminder))
			return result, err
		}
	}

	if err := uow.Commit(ctx); err != nil {
		logging.Error(ctx, s.log, err, logging.Entry("input", input), logging.Entry("reminder", createdReminder))
		return result, err
	}

	s.log.Info(
		ctx,
		"Reminder successfully created.",
		logging.Entry("reminder", createdReminder),
	)
	result.Reminder.FromReminderAndChannels(createdReminder, channelIDs)
	return result, nil
}

func (s *service) readChannels(
	ctx context.Context,
	uow uow.Context,
	input Input,
) ([]channel.ID, error) {
	if len(input.ChannelIDs) == 0 {
		return nil, reminder.ErrReminderChannelsNotSet
	}
	channelIDs := make([]channel.ID, 0, len(input.ChannelIDs))
	for channelID := range input.ChannelIDs {
		channelIDs = append(channelIDs, channelID)
	}

	channels, err := uow.Channels().Read(
		ctx,
		channel.ReadOptions{
			IDIn:         c.NewOptional(channelIDs, true),
			UserIDEquals: c.NewOptional(input.UserID, true),
		},
	)
	if err != nil {
		logging.Error(ctx, s.log, err, logging.Entry("input", input))
		return nil, err
	}
	readChannelIDs := make(map[channel.ID]struct{})
	resultChannelIDs := make([]channel.ID, 0, len(channelIDs))
	for _, readChannel := range channels {
		if !readChannel.IsVerified() {
			return nil, reminder.ErrReminderChannelsNotVerified
		}
		readChannelIDs[readChannel.ID] = struct{}{}
		resultChannelIDs = append(resultChannelIDs, readChannel.ID)
	}
	for _, channelID := range channelIDs {
		_, ok := readChannelIDs[channelID]
		if !ok {
			return nil, reminder.ErrReminderChannelsNotValid
		}
	}

	return resultChannelIDs, nil
}

func (s *service) checkUserLimits(ctx context.Context, uow uow.Context, limits user.Limits, input Input) error {
	if limits.ReminderEveryPerDayCount.IsPresent && input.Every.IsPresent {
		if input.Every.Value.PerDayCount() > limits.ReminderEveryPerDayCount.Value {
			return user.ErrLimitReminderEveryPerDayCountExceeded
		}
	}

	if limits.ActiveReminderCount.IsPresent {
		activeReminderCount, err := uow.Reminders().Count(
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
			logging.Error(ctx, s.log, err, logging.Entry("input", input))
			return err
		}
		if activeReminderCount >= uint(limits.ActiveReminderCount.Value) {
			return user.ErrLimitActiveReminderCountExceeded
		}
	}

	now := s.now()
	if limits.MonthlySentReminderCount.IsPresent && now.Year() == input.At.Year() && now.Month() == input.At.Month() {
		sentAfter := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
		sentReminderCount, err := uow.Reminders().Count(
			ctx,
			reminder.ReadOptions{
				CreatedByEquals: c.NewOptional(input.UserID, true),
				StatusIn:        c.NewOptional([]reminder.Status{reminder.StatusSentSuccess}, true),
				SentAfter:       c.NewOptional(sentAfter, true),
			},
		)
		if err != nil {
			logging.Error(ctx, s.log, err, logging.Entry("input", input), logging.Entry("sentAfter", sentAfter))
			return err
		}
		if sentReminderCount >= uint(limits.MonthlySentReminderCount.Value) {
			return user.ErrLimitSentReminderCountExceeded
		}
	}

	return nil
}
