package channel

type Type struct {
	v string
}

var (
	Websocket = Type{v: "websocket"}
	Telegram  = Type{v: "telegram"}
	Email     = Type{v: "email"}
)
