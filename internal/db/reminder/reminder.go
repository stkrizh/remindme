package remidner

import (
	"context"
	"database/sql"
	"fmt"
	e "remindme/internal/core/domain/errors"
	"remindme/internal/core/domain/reminder"
	"remindme/internal/core/domain/user"
	"remindme/internal/db/sqlcgen"
)

type PgxReminderRepository struct {
	queries *sqlcgen.Queries
}

func NewPgxReminderRepository(db sqlcgen.DBTX) *PgxReminderRepository {
	if db == nil {
		panic(e.NewNilArgumentError("db"))
	}
	return &PgxReminderRepository{queries: sqlcgen.New(db)}
}

func (r *PgxReminderRepository) Create(
	ctx context.Context,
	input reminder.CreateInput,
) (rem reminder.Reminder, err error) {
	dbReminder, err := r.queries.CreateReminder(
		ctx,
		sqlcgen.CreateReminderParams{
			UserID:    int64(input.CreatedBy),
			CreatedAt: input.CreatedAt,
			At:        input.At,
			Every: sql.NullString{
				String: input.Every.Value.String(),
				Valid:  input.Every.IsPresent,
			},
			ScheduledAt: sql.NullTime{
				Time:  input.ScheduledAt.Value,
				Valid: input.ScheduledAt.IsPresent,
			},
			SentAt: sql.NullTime{
				Time:  input.SentAt.Value,
				Valid: input.SentAt.IsPresent,
			},
			CanceledAt: sql.NullTime{
				Time:  input.CanceledAt.Value,
				Valid: input.CanceledAt.IsPresent,
			},
			Status: input.Status.String(),
		},
	)
	if err != nil {
		return rem, err
	}
	return decodeReminder(dbReminder)
}

func (r *PgxReminderRepository) Read(
	ctx context.Context,
	options reminder.ReadOptions,
) (reminders []reminder.Reminder, err error) {
	var statusIn []string
	if options.StatusIn.IsPresent {
		statusIn = make([]string, len(options.StatusIn.Value))
		for ix, status := range options.StatusIn.Value {
			statusIn[ix] = status.String()
		}
	}

	dbReminders, err := r.queries.ReadReminders(
		ctx,
		sqlcgen.ReadRemindersParams{
			AnyUserID:     !options.CreatedByEquals.IsPresent,
			UserIDEquals:  int64(options.CreatedByEquals.Value),
			AnySentAt:     !options.SentAfter.IsPresent,
			SentAfter:     options.SentAfter.Value,
			AnyStatus:     !options.StatusIn.IsPresent,
			StatusIn:      statusIn,
			OrderByIDAsc:  options.OrderBy == reminder.OrderByIDAsc,
			OrderByIDDesc: options.OrderBy == reminder.OrderByIDDesc,
			OrderByAtAsc:  options.OrderBy == reminder.OrderByAtAsc,
			OrderByAtDesc: options.OrderBy == reminder.OrderByAtDesc,
		},
	)
	if err != nil {
		return reminders, err
	}

	reminders = make([]reminder.Reminder, len(dbReminders))
	for ix, dbReminder := range dbReminders {
		rem, err := decodeReminder(dbReminder)
		if err != nil {
			return reminders, err
		}
		reminders[ix] = rem
	}

	return reminders, nil
}

func (r *PgxReminderRepository) Count(ctx context.Context, options reminder.ReadOptions) (uint, error) {
	var statusIn []string
	if options.StatusIn.IsPresent {
		statusIn = make([]string, len(options.StatusIn.Value))
		for ix, status := range options.StatusIn.Value {
			statusIn[ix] = status.String()
		}
	}

	count, err := r.queries.CountReminders(
		ctx,
		sqlcgen.CountRemindersParams{
			AnyUserID:    !options.CreatedByEquals.IsPresent,
			UserIDEquals: int64(options.CreatedByEquals.Value),
			AnySentAt:    !options.SentAfter.IsPresent,
			SentAfter:    options.SentAfter.Value,
			AnyStatus:    !options.StatusIn.IsPresent,
			StatusIn:     statusIn,
		},
	)
	if err != nil {
		return 0, err
	}

	return uint(count), nil
}

func decodeReminder(dbReminder sqlcgen.Reminder) (rem reminder.Reminder, err error) {
	rem.ID = reminder.ID(dbReminder.ID)
	rem.CreatedBy = user.ID(dbReminder.UserID)
	rem.At = dbReminder.At
	if dbReminder.Every.Valid {
		every, err := reminder.ParseEvery(dbReminder.Every.String)
		if err != nil {
			return rem, err
		}
		rem.Every.Value = every
		rem.Every.IsPresent = true
	}
	status, err := reminder.ParseStatus(dbReminder.Status)
	if err != nil {
		return rem, err
	}
	rem.Status = status
	rem.CreatedAt = dbReminder.CreatedAt
	if dbReminder.ScheduledAt.Valid {
		rem.ScheduledAt.Value = dbReminder.ScheduledAt.Time
		rem.ScheduledAt.IsPresent = true
	}
	if dbReminder.SentAt.Valid {
		rem.SentAt.Value = dbReminder.SentAt.Time
		rem.SentAt.IsPresent = true
	}
	if dbReminder.CanceledAt.Valid {
		rem.CanceledAt.Value = dbReminder.CanceledAt.Time
		rem.CanceledAt.IsPresent = true
	}
	return rem, rem.Validate()
}

type PgxReminderChannelRepository struct {
	queries *sqlcgen.Queries
}

func NewPgxReminderChannelRepository(db sqlcgen.DBTX) *PgxReminderChannelRepository {
	if db == nil {
		panic(e.NewNilArgumentError("db"))
	}
	return &PgxReminderChannelRepository{queries: sqlcgen.New(db)}
}

func (r *PgxReminderChannelRepository) Create(
	ctx context.Context,
	input reminder.CreateChannelsInput,
) (reminder.ChannelIDs, error) {
	if len(input.ChannelIDs) == 0 {
		return input.ChannelIDs, nil
	}
	params := make([]sqlcgen.CreateReminderChannelsParams, 0, len(input.ChannelIDs))
	for channelID := range input.ChannelIDs {
		params = append(
			params,
			sqlcgen.CreateReminderChannelsParams{
				ReminderID: int64(input.ReminderID),
				ChannelID:  int64(channelID),
			},
		)
	}
	createdCount, err := r.queries.CreateReminderChannels(ctx, params)
	if err != nil {
		return input.ChannelIDs, err
	}
	if createdCount != int64(len(input.ChannelIDs)) {
		return input.ChannelIDs, fmt.Errorf(
			"could not create some reminder channels, count %d, created %d",
			len(input.ChannelIDs),
			createdCount,
		)
	}
	return input.ChannelIDs, nil
}
