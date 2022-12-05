package user

import (
	"fmt"
	c "remindme/internal/domain/common"
	e "remindme/internal/domain/errors"
	"time"
)

type ID int64

type Email string

type PasswordHash string

func (p PasswordHash) String() string {
	return "***"
}

type RawPassword string

func (p RawPassword) String() string {
	return "***"
}

type Identity string

type ActivationToken string

type SessionToken string

type User struct {
	ID              ID
	Email           c.Optional[Email]
	PasswordHash    c.Optional[PasswordHash]
	Identity        c.Optional[Identity]
	CreatedAt       time.Time
	ActivatedAt     c.Optional[time.Time]
	ActivationToken c.Optional[ActivationToken]
	LastLoginAt     c.Optional[time.Time]
}

func (u *User) Validate() error {
	if u.Email.IsPresent {
		if !u.PasswordHash.IsPresent {
			return e.NewInvalidStateError(fmt.Sprintf("password hash is not set for user %d", u.ID))
		}
		return nil
	}
	if !u.Identity.IsPresent {
		return e.NewInvalidStateError(fmt.Sprintf("neither email nor identity is not defined for user %d", u.ID))
	}
	return nil
}

func (u *User) IsAnonymous() bool {
	if u.Email.IsPresent && u.PasswordHash.IsPresent {
		return false
	}
	if u.Identity.IsPresent {
		return true
	}
	panic(fmt.Sprintf("neither email nor identity is not defined for user %d", u.ID))
}
