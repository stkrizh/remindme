package reminder

import "errors"

var ErrParseStatus = errors.New("invalid status")

type Status (string)

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

const (
	StatusInvalid           = Status("")
	StatusCreated           = Status("created")
	StatusScheduled         = Status("scheduled")
	StatusSending           = Status("sending")
	StatusSentSuccess       = Status("sent_success")
	StatusSentError         = Status("sent_error")
	StatusSentLimitExceeded = Status("sent_limit_exceeded")
	StatusCanceled          = Status("canceled")
)
