package user

import (
	"context"
	"database/sql"
	c "remindme/internal/core/domain/common"
	e "remindme/internal/core/domain/errors"
	"remindme/internal/core/domain/user"
	"remindme/internal/db/sqlcgen"
)

type PgxLimitsRepository struct {
	queries *sqlcgen.Queries
}

func NewPgxLimitsRepository(db sqlcgen.DBTX) *PgxLimitsRepository {
	if db == nil {
		panic(e.NewNilArgumentError("db"))
	}
	return &PgxLimitsRepository{queries: sqlcgen.New(db)}
}

func (r *PgxLimitsRepository) Create(ctx context.Context, input user.CreateLimitsInput) (user.Limits, error) {
	dbLimits, err := r.queries.CreateLimits(ctx, sqlcgen.CreateLimitsParams{
		UserID: int64(input.UserID),
		EmailChannelCount: sql.NullInt32{
			Int32: int32(input.Limits.EmailChannelCount.Value),
			Valid: input.Limits.EmailChannelCount.IsPresent,
		},
		TelegramChannelCount: sql.NullInt32{
			Int32: int32(input.Limits.TelegramChannelCount.Value),
			Valid: input.Limits.TelegramChannelCount.IsPresent,
		},
		ActiveReminderCount: sql.NullInt32{
			Int32: int32(input.Limits.ActiveReminderCount.Value),
			Valid: input.Limits.ActiveReminderCount.IsPresent,
		},
		MonthlySentReminderCount: sql.NullInt32{
			Int32: int32(input.Limits.MonthlySentReminderCount.Value),
			Valid: input.Limits.MonthlySentReminderCount.IsPresent,
		},
		ReminderEveryPerDayCount: sql.NullFloat64{
			Float64: input.Limits.ReminderEveryPerDayCount.Value,
			Valid:   input.Limits.ReminderEveryPerDayCount.IsPresent,
		},
	})
	return decodeLimits(dbLimits), err
}

func (r *PgxLimitsRepository) GetUserLimits(ctx context.Context, userID user.ID) (user.Limits, error) {
	dbLimits, err := r.queries.GetUserLimits(ctx, int64(userID))
	return decodeLimits(dbLimits), err
}

func (r *PgxLimitsRepository) GetUserLimitsWithLock(ctx context.Context, userID user.ID) (user.Limits, error) {
	dbLimits, err := r.queries.GetUserLimitsWithLock(ctx, int64(userID))
	return decodeLimits(dbLimits), err
}

func decodeLimits(l sqlcgen.Limit) user.Limits {
	return user.Limits{
		EmailChannelCount:    c.NewOptional(uint32(l.EmailChannelCount.Int32), l.EmailChannelCount.Valid),
		TelegramChannelCount: c.NewOptional(uint32(l.TelegramChannelCount.Int32), l.TelegramChannelCount.Valid),
		ActiveReminderCount:  c.NewOptional(uint32(l.ActiveReminderCount.Int32), l.ActiveReminderCount.Valid),
		MonthlySentReminderCount: c.NewOptional(
			uint32(l.MonthlySentReminderCount.Int32),
			l.MonthlySentReminderCount.Valid,
		),
		ReminderEveryPerDayCount: c.NewOptional(
			l.ReminderEveryPerDayCount.Float64,
			l.ReminderEveryPerDayCount.Valid,
		),
	}
}
