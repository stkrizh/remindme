package channel

import "errors"

var (
	ErrChannelDoesNotExist     = errors.New("channel does not exist")
	ErrInvalidVerificationData = errors.New("invalid channel verification data")
)
