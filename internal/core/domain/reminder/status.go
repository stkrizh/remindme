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
	case "sending":
		return StatusSending, nil
	case "sent_success":
		return StatusSentSuccess, nil
	case "sent_error":
		return StatusSentError, nil
	case "sent_limit_exceeded":
		return StatusSentLimitExceeded, nil
	case "canceled":
		return StatusCanceled, nil
	default:
		return StatusInvalid, ErrParseStatus
	}
}

var (
	StatusInvalid           = Status{}
	StatusCreated           = Status{v: "created"}
	StatusScheduled         = Status{v: "scheduled"}
	StatusSending           = Status{v: "sending"}
	StatusSentSuccess       = Status{v: "sent_success"}
	StatusSentError         = Status{v: "sent_error"}
	StatusSentLimitExceeded = Status{v: "sent_limit_exceeded"}
	StatusCanceled          = Status{v: "canceled"}
)
