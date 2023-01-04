package reminder

import "errors"

var ErrParseStatus = errors.New("invalid status")

type Status struct {
	v string
}

func (s Status) String() string {
	return s.v
}

func ParseStatus(value string) (Status, error) {
	switch value {
	case "created":
		return StatusCreated, nil
	case "scheduled":
		return StatusScheduled, nil
	case "send_success":
		return StatusSendSuccess, nil
	case "send_error":
		return StatusSendError, nil
	case "send_limit_exceeded":
		return StatusSendLimitExceeded, nil
	case "canceled":
		return StatusCanceled, nil
	default:
		return StatusUnknown, ErrParseStatus
	}
}

var (
	StatusUnknown           = Status{}
	StatusCreated           = Status{v: "created"}
	StatusScheduled         = Status{v: "scheduled"}
	StatusSendSuccess       = Status{v: "send_success"}
	StatusSendError         = Status{v: "send_error"}
	StatusSendLimitExceeded = Status{v: "send_limit_exceeded"}
	StatusCanceled          = Status{v: "canceled"}
)
