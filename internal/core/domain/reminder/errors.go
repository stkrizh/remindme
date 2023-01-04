package reminder

import "errors"

var (
	ErrReminderTooEarly        = errors.New("reminder time is too early")
	ErrReminderTooManyChannels = errors.New("reminder has too many channels")
	ErrReminderChannelsNotSet  = errors.New("reminder channels are not set")
	ErrReminderInvalidChannels = errors.New("reminder channels are not valid")
)
