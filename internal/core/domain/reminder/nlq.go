package reminder

import (
	"context"
	"time"
)

type NaturalLanguageQueryParser interface {
	Parse(ctx context.Context, query string, userLocalTime time.Time) (CreateReminderParams, error)
}
