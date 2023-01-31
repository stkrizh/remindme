package updatereminderchannels

import (
	"context"
	"errors"
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
	ReminderID reminder.ID
	ChannelIDs reminder.ChannelIDs
}

func (i Input) WithAuthenticatedUser(u user.User) auth.Input {
	i.UserID = u.ID
	return i
}

func (i Input) Validate() error {
	if len(i.ChannelIDs) == 0 {
		return reminder.ErrReminderChannelsNotSet
	}
	if len(i.ChannelIDs) > reminder.MAX_CHANNEL_COUNT {
		return reminder.ErrReminderTooManyChannels
	}
	return nil
}

type Result struct {
	ChannelIDs []channel.ID
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
	err = input.Validate()
	if err != nil {
		return result, err
	}

	uow, err := s.unitOfWork.Begin(ctx)
	if err != nil {
		logging.Error(ctx, s.log, err, logging.Entry("input", input))
	}
	defer uow.Rollback(ctx)

	reminderRepo := uow.Reminders()
	reminderRepo.Lock(ctx, input.ReminderID)
	rem, err := reminderRepo.GetByID(ctx, input.ReminderID)
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
		s.log.Info(ctx, "Reminder is not active, channels can't be updated.", logging.Entry("input", input))
		return result, reminder.ErrReminderNotActive
	}

	channelIDs, err := s.readChannels(ctx, uow, input)
	if err != nil {
		return result, err
	}

	reminderChannelRepo := uow.ReminderChannels()
	err = reminderChannelRepo.DeleteByReminderID(ctx, input.ReminderID)
	if err != nil {
		logging.Error(ctx, s.log, err, logging.Entry("input", input))
		return result, err
	}
	_, err = reminderChannelRepo.Create(
		ctx,
		reminder.NewCreateChannelsInput(input.ReminderID, channelIDs...),
	)
	if err != nil {
		logging.Error(ctx, s.log, err, logging.Entry("input", input))
		return result, err
	}

	if err := uow.Commit(ctx); err != nil {
		logging.Error(ctx, s.log, err, logging.Entry("input", input))
		return result, err
	}

	s.log.Info(ctx, "Reminder channels have been successfully updated.", logging.Entry("input", input))
	return Result{ChannelIDs: channelIDs}, nil
}

func (s *service) readChannels(
	ctx context.Context,
	uow uow.Context,
	input Input,
) ([]channel.ID, error) {
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
