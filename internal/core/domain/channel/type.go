package channel

type Type struct {
	v string
}

func ParseType(t string) Type {
	switch t {
	case "internal":
		return Internal
	case "telegram":
		return Telegram
	case "email":
		return Email
	default:
		return Unknown
	}
}

func (t Type) String() string {
	return t.v
}

var (
	Unknown  = Type{}
	Internal = Type{v: "internal"}
	Telegram = Type{v: "telegram"}
	Email    = Type{v: "email"}
)
