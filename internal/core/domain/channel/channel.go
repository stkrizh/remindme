package channel

import (
	c "remindme/internal/core/domain/common"
	e "remindme/internal/core/domain/errors"
	"remindme/internal/core/domain/user"
	"time"
)

type ID int64

type VerificationToken string

type Channel struct {
	ID                ID
	Type              Type
	Settings          Settings
	CreatedBy         user.ID
	CreatedAt         time.Time
	VerificationToken c.Optional[VerificationToken]
	VerifiedAt        c.Optional[time.Time]
}

func (c *Channel) Validate() error {
	if !(c.VerifiedAt.IsPresent || c.VerificationToken.IsPresent) {
		return e.NewInvalidStateError("either VerifiedAt or VerificationToken must be defined")
	}
	if c.Type == Unknown {
		return e.NewInvalidStateError("invalid channel type")
	}
	return nil
}

func (c *Channel) IsVerified() bool {
	return c.VerifiedAt.IsPresent
}
