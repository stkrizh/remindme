package cancelreminder

import (
	"context"
	c "remindme/internal/core/domain/common"
	e "remindme/internal/core/domain/errors"
	"remindme/internal/core/domain/logging"
	"remindme/internal/core/domain/reminder"
	"remindme/internal/core/domain/user"
	"remindme/internal/core/services"
	"remindme/internal/core/services/auth"
	"time"
)

type Input struct {
	UserID     user.ID
	At         time.Time
	Every      c.Optional[reminder.Every]
	ChannelIDs reminder.ChannelIDs
}

func (i Input) WithAuthenticatedUser(u user.User) auth.Input {
	i.UserID = u.ID
	return i
}

type Result struct{}

type service struct {
	log                logging.Logger
	reminderRepository reminder.ReminderRepository
	now                func() time.Time
}

func New(
	log logging.Logger,
	reminderRepository reminder.ReminderRepository,
	now func() time.Time,
) services.Service[Input, Result] {
	if log == nil {
		panic(e.NewNilArgumentError("log"))
	}
	if reminderRepository == nil {
		panic(e.NewNilArgumentError("reminderRepository"))
	}
	if now == nil {
		panic(e.NewNilArgumentError("now"))
	}
	return &service{
		log:                log,
		reminderRepository: reminderRepository,
		now:                now,
	}
}

func (s *service) Run(ctx context.Context, input Input) (result Result, err error) {
	return result, err
}
