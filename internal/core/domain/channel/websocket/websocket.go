package websocket

import "remindme/internal/core/domain/channel"

type Settings struct{}

func (s Settings) Type() channel.Type {
	return channel.WEBSOCKET
}
