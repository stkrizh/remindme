package reminder

import "context"

type Sender interface {
	SendReminder(ctx context.Context, reminder ReminderWithChannels) error
}
