package email

import (
	"remindme/internal/core/domain/channel"
	c "remindme/internal/core/domain/common"
)

type Settings struct {
	Email c.Email
}

func (s Settings) Type() channel.Type {
	return channel.EMAIL
}
