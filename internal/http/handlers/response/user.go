package response

import (
	"remindme/internal/core/domain/user"
	"time"
)

type User struct {
	ID          int64     `json:"id"`
	Email       *string   `json:"email,omitempty"`
	Identity    *string   `json:"identity,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	ActivatedAt time.Time `json:"activated_at"`
}

func (u *User) FromDomainUser(du user.User) {
	u.ID = int64(du.ID)
	if du.Email.IsPresent {
		email := string(du.Email.Value)
		u.Email = &email
	}
	if du.Identity.IsPresent {
		identity := string(du.Identity.Value)
		u.Identity = &identity
	}
	u.CreatedAt = du.CreatedAt
	u.ActivatedAt = du.ActivatedAt.Value
}
