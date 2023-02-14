package reminder

import (
	"context"
	"time"
)

type NaturalLanguageQueryParser interface {
	Parse(ctx context.Context, query string, loc *time.Location) (CreateReminderParams, error)
}
