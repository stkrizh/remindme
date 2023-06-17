package channel

type Type (string)

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

const (
	Unknown  = Type("")
	Internal = Type("internal")
	Telegram = Type("telegram")
	Email    = Type("email")
)
