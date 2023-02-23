package createreminderbynlq

import (
	"context"
	"errors"
	"remindme/internal/core/domain/channel"
	c "remindme/internal/core/domain/common"
	e "remindme/internal/core/domain/errors"
	"remindme/internal/core/domain/logging"
	"remindme/internal/core/domain/reminder"
	"remindme/internal/core/domain/user"
	"remindme/internal/core/services"
	"remindme/internal/core/services/auth"
	createreminder "remindme/internal/core/services/create_reminder"
	"time"
)

type Input struct {
	User  user.User
	Query string
}

func (i Input) WithAuthenticatedUser(u user.User) auth.Input {
	i.User = u
	return i
}

type service struct {
	log           logging.Logger
	parser        reminder.NaturalLanguageQueryParser
	channelRepo   channel.Repository
	now           func() time.Time
	createService services.Service[createreminder.Input, createreminder.Result]
}

func New(
	log logging.Logger,
	parser reminder.NaturalLanguageQueryParser,
	channelRepo channel.Repository,
	now func() time.Time,
	createSerservice services.Service[createreminder.Input, createreminder.Result],
) services.Service[Input, createreminder.Result] {
	if log == nil {
		panic(e.NewNilArgumentError("log"))
	}
	if parser == nil {
		panic(e.NewNilArgumentError("parser"))
	}
	if channelRepo == nil {
		panic(e.NewNilArgumentError("channelRepo"))
	}
	if now == nil {
		panic(e.NewNilArgumentError("now"))
	}
	if createSerservice == nil {
		panic(e.NewNilArgumentError("createSerservice"))
	}

	return &service{
		log:           log,
		parser:        parser,
		channelRepo:   channelRepo,
		now:           now,
		createService: createSerservice,
	}
}

func (s *service) Run(ctx context.Context, input Input) (result createreminder.Result, err error) {
	defaultChannel, err := s.getDefaultChannel(ctx, input.User.ID)
	if err != nil {
		logging.Error(ctx, s.log, err, logging.Entry("input", input))
		return result, err
	}
	s.log.Info(
		ctx,
		"Found default channel for reminder.",
		logging.Entry("channelID", defaultChannel.ID),
	)
	createParams, err := s.parser.Parse(ctx, input.Query, s.now().In(input.User.TimeZone))
	if err != nil {
		switch {
		case errors.Is(err, reminder.ErrNaturalQueryParsing):
			// do nothing
		default:
			logging.Error(ctx, s.log, err, logging.Entry("input", input))
		}
		return result, err
	}
	s.log.Info(
		ctx,
		"Reminder creating params parsed.",
		logging.Entry("userID", input.User.ID),
		logging.Entry("query", input.Query),
		logging.Entry("params", createParams),
	)
	result, err = s.createService.Run(ctx, createreminder.Input{
		UserID:     input.User.ID,
		At:         createParams.At,
		Body:       createParams.Body,
		Every:      createParams.Every,
		ChannelIDs: reminder.NewChannelIDs(defaultChannel.ID),
	})
	return result, err
}

func (s *service) getDefaultChannel(ctx context.Context, userID user.ID) (defChannel channel.Channel, err error) {
	channels, err := s.channelRepo.Read(ctx, channel.ReadOptions{
		UserIDEquals: c.NewOptional(userID, true),
		TypeEquals:   c.NewOptional(channel.Websocket, true),
		OrderBy:      channel.OrderByIDAsc,
		Limit:        c.NewOptional(uint(1), true),
	})
	if err != nil {
		return defChannel, err
	}
	return channels[0], err
}
