package reminder

import "errors"

var (
	ErrReminderAtTimeIsNotUTC      = errors.New("remider at time is not UTC")
	ErrReminderTooEarly            = errors.New("reminder time is too early")
	ErrReminderTooLate             = errors.New("reminder time is too late")
	ErrReminderTooManyChannels     = errors.New("reminder has too many channels")
	ErrReminderChannelsNotSet      = errors.New("reminder channels are not set")
	ErrReminderChannelsNotValid    = errors.New("reminder channels are not valid")
	ErrReminderChannelsNotVerified = errors.New("reminder channels are not verified")
	ErrReminderDoesNotExist        = errors.New("reminder does not exist")
	ErrReminderPermission          = errors.New("reminder permission error")
	ErrReminderNotActive           = errors.New("reminder is not active")
	ErrReminderIsSending           = errors.New("reminder is sending")
)
