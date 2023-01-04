package reminder

import "context"

type Scheduler interface {
	ScheduleReminder(ctx context.Context, r Reminder) error
}
