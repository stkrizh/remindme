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
	userLocalTime := s.now().In(input.User.TimeZone)
	createParams, err := s.parser.Parse(ctx, input.Query, userLocalTime)
	if err != nil {
		switch {
		case errors.Is(err, reminder.ErrNaturalQueryParsing):
			s.log.Info(
				ctx,
				"Could not parse reminder creation params.",
				logging.Entry("query", input.Query),
				logging.Entry("err", err),
			)
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
	channelIDs, err := s.getChannelIDs(ctx, input.User.ID, createParams.At.Sub(userLocalTime))
	if err != nil {
		logging.Error(ctx, s.log, err, logging.Entry("input", input))
		return result, err
	}
	s.log.Info(
		ctx,
		"Found channels for sending reminder.",
		logging.Entry("channelIDs", channelIDs),
	)
	result, err = s.createService.Run(ctx, createreminder.Input{
		UserID:     input.User.ID,
		At:         createParams.At.In(time.UTC),
		Body:       createParams.Body,
		Every:      createParams.Every,
		ChannelIDs: channelIDs,
	})
	return result, err
}

func (s *service) getChannelIDs(
	ctx context.Context,
	userID user.ID,
	reminderWillBeSentAfter time.Duration,
) (channelIDs reminder.ChannelIDs, err error) {
	channels, err := s.channelRepo.Read(
		ctx,
		channel.ReadOptions{
			UserIDEquals: c.NewOptional(userID, true),
			OrderBy:      channel.OrderByIDDesc,
		},
	)
	if err != nil {
		return channelIDs, err
	}

	channelIDs = make(reminder.ChannelIDs)
	var emailChannelID channel.ID
	var tlgChannelID channel.ID
	var wsChannelID channel.ID
	for _, ch := range channels {
		if !ch.IsVerified() {
			continue
		}
		if ch.Type == channel.Email && emailChannelID == 0 {
			emailChannelID = ch.ID
			continue
		}
		if ch.Type == channel.Telegram && tlgChannelID == 0 {
			tlgChannelID = ch.ID
			continue
		}
		if ch.Type == channel.Websocket && wsChannelID == 0 {
			wsChannelID = ch.ID
			continue
		}
	}

	if emailChannelID == 0 && tlgChannelID == 0 && wsChannelID != 0 {
		channelIDs[wsChannelID] = struct{}{}
		return channelIDs, nil
	}

	if reminderWillBeSentAfter <= time.Hour && wsChannelID != 0 {
		channelIDs[wsChannelID] = struct{}{}
	}
	if reminderWillBeSentAfter > time.Hour && emailChannelID != 0 {
		channelIDs[emailChannelID] = struct{}{}
	}
	if tlgChannelID != 0 {
		channelIDs[tlgChannelID] = struct{}{}
	}
	return channelIDs, nil
}
