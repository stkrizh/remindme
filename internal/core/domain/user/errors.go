package user

import (
	"errors"
)

var (
	ErrEmailAlreadyExists         = errors.New("email already exists")
	ErrUserDoesNotExist           = errors.New("user does not exist")
	ErrInvalidCredentials         = errors.New("invalid credentials")
	ErrUserIsNotActive            = errors.New("user is not active")
	ErrSessionDoesNotExist        = errors.New("session does not exist")
	ErrInvalidPasswordFResetToken = errors.New("invalid password reset token")
	ErrInvalidActivationToken     = errors.New("invalid activation token")
)

var (
	ErrLimitEmailChannelCountExceeded        = errors.New("email channel count limit exceeded")
	ErrLimitTelegramChannelCountExceeded     = errors.New("telegram channel count limit exceeded")
	ErrLimitReminderEveryPerDayCountExceeded = errors.New("reminder every per day count exceeded")
	ErrLimitActiveReminderCountExceeded      = errors.New("active reminder count limit exceeded")
	ErrLimitSentReminderCountExceeded        = errors.New("sent reminder count monthly limit exceeded")
)
