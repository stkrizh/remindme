package remidner

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"remindme/internal/core/domain/channel"
	e "remindme/internal/core/domain/errors"
	"remindme/internal/core/domain/reminder"
	"remindme/internal/core/domain/user"
	"remindme/internal/db/sqlcgen"
	"time"

	"github.com/jackc/pgx/v4"
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
			Status: string(input.Status),
			Body:   input.Body,
		},
	)
	if err != nil {
		return rem, err
	}
	return decodeReminder(dbReminder)
}

func (r *PgxReminderRepository) Lock(ctx context.Context, id reminder.ID) error {
	// The method works only within a DB transaction
	return r.queries.LockReminder(ctx, int64(id))
}

func (r *PgxReminderRepository) GetByID(
	ctx context.Context,
	id reminder.ID,
) (rem reminder.ReminderWithChannels, err error) {
	dbReminder, err := r.queries.GetReminderByID(ctx, int64(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return rem, reminder.ErrReminderDoesNotExist
		}
		return rem, err
	}
	return decodeReminderWithChannels(dbReminder)
}

func (r *PgxReminderRepository) Read(
	ctx context.Context,
	options reminder.ReadOptions,
) (reminders []reminder.ReminderWithChannels, err error) {
	var statusIn []string
	if options.StatusIn.IsPresent {
		statusIn = make([]string, len(options.StatusIn.Value))
		for ix, status := range options.StatusIn.Value {
			statusIn[ix] = string(status)
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
			AllRows:       !options.Limit.IsPresent,
			Limit:         int32(options.Limit.Value),
			Offset:        int32(options.Offset),
		},
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return reminders, nil
		}
		return reminders, err
	}

	reminders = make([]reminder.ReminderWithChannels, 0, len(dbReminders))
	for _, dbReminder := range dbReminders {
		rem, err := decodeReminderWithChannels(dbReminder)
		if err != nil {
			return reminders, err
		}
		reminders = append(reminders, rem)
	}

	return reminders, nil
}

func (r *PgxReminderRepository) Count(ctx context.Context, options reminder.ReadOptions) (uint, error) {
	var statusIn []string
	if options.StatusIn.IsPresent {
		statusIn = make([]string, len(options.StatusIn.Value))
		for ix, status := range options.StatusIn.Value {
			statusIn[ix] = string(status)
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

func (r *PgxReminderRepository) Update(
	ctx context.Context,
	input reminder.UpdateInput,
) (rem reminder.Reminder, err error) {
	dbReminder, err := r.queries.UpdateReminder(
		ctx,
		sqlcgen.UpdateReminderParams{
			ID:            int64(input.ID),
			DoAtUpdate:    input.DoAtUpdate,
			At:            input.At,
			DoBodyUpdate:  input.DoBodyUpdate,
			Body:          input.Body,
			DoEveryUpdate: input.DoEveryUpdate,
			Every: sql.NullString{
				Valid:  input.Every.IsPresent,
				String: input.Every.Value.String(),
			},
			DoStatusUpdate:      input.DoStatusUpdate,
			Status:              string(input.Status),
			DoScheduledAtUpdate: input.DoScheduledAtUpdate,
			ScheduledAt: sql.NullTime{
				Valid: input.ScheduledAt.IsPresent,
				Time:  input.ScheduledAt.Value,
			},
			DoSentAtUpdate: input.DoSentAtUpdate,
			SentAt: sql.NullTime{
				Valid: input.SentAt.IsPresent,
				Time:  input.SentAt.Value,
			},
			DoCanceledAtUpdate: input.DoCanceledAtUpdate,
			CanceledAt: sql.NullTime{
				Valid: input.CanceledAt.IsPresent,
				Time:  input.CanceledAt.Value,
			},
		},
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return rem, reminder.ErrReminderDoesNotExist
		}
		return rem, err
	}
	return decodeReminder(dbReminder)
}

func (r *PgxReminderRepository) Schedule(
	ctx context.Context,
	input reminder.ScheduleInput,
) (reminders []reminder.Reminder, err error) {
	dbReminders, err := r.queries.ScheduleReminders(ctx, sqlcgen.ScheduleRemindersParams{
		Status:                string(reminder.StatusScheduled),
		AtBefore:              input.AtBefore,
		ScheduledAt:           input.ScheduledAt,
		StatusesForScheduling: []string{string(reminder.StatusCreated)},
	})
	if err != nil {
		return reminders, err
	}
	reminders = make([]reminder.Reminder, 0, len(dbReminders))
	for _, dbReminder := range dbReminders {
		rem, err := decodeReminder(dbReminder)
		if err != nil {
			return reminders, err
		}
		reminders = append(reminders, rem)
	}

	return reminders, nil
}

func (r *PgxReminderRepository) Delete(
	ctx context.Context,
	id reminder.ID,
) error {
	_, err := r.queries.DeleteReminder(ctx, int64(id))
	if errors.Is(err, pgx.ErrNoRows) {
		return reminder.ErrReminderDoesNotExist
	}
	return err
}

func decodeReminder(dbReminder sqlcgen.Reminder) (rem reminder.Reminder, err error) {
	rem.ID = reminder.ID(dbReminder.ID)
	rem.CreatedBy = user.ID(dbReminder.UserID)
	rem.At = dbReminder.At
	rem.Body = dbReminder.Body
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

func decodeReminderWithChannels(dbRow struct {
	ID          int64
	UserID      int64
	CreatedAt   time.Time
	At          time.Time
	Body        string
	Status      string
	Every       sql.NullString
	ScheduledAt sql.NullTime
	SentAt      sql.NullTime
	CanceledAt  sql.NullTime
	ChannelIds  []int64
}) (rem reminder.ReminderWithChannels, err error) {
	dbReminder := sqlcgen.Reminder{
		ID:          dbRow.ID,
		UserID:      dbRow.UserID,
		CreatedAt:   dbRow.CreatedAt,
		At:          dbRow.At,
		Body:        dbRow.Body,
		Status:      dbRow.Status,
		Every:       dbRow.Every,
		ScheduledAt: dbRow.ScheduledAt,
		SentAt:      dbRow.SentAt,
		CanceledAt:  dbRow.CanceledAt,
	}
	r, err := decodeReminder(dbReminder)
	if err != nil {
		return rem, err
	}

	channelIDs := make([]channel.ID, 0, len(dbRow.ChannelIds))
	for _, rawChannelID := range dbRow.ChannelIds {
		channelIDs = append(channelIDs, channel.ID(rawChannelID))
	}

	rem.FromReminderAndChannels(r, channelIDs)
	return rem, nil
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

func (r *PgxReminderChannelRepository) DeleteByReminderID(
	ctx context.Context,
	reminderID reminder.ID,
) error {
	return r.queries.DeleteReminderChannels(ctx, int64(reminderID))
}
