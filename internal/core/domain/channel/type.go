package channel

type Type struct {
	v string
}

var (
	WEBSOCKET = Type{v: "websocket"}
	TELEGRAM  = Type{v: "telegram"}
	EMAIL     = Type{v: "email"}
)
