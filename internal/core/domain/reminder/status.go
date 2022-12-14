package reminder

type Status struct {
	v string
}

var (
	SCHEDULED = Status{v: "scheduled"}
	SENT      = Status{v: "sent"}
	CANCELED  = Status{v: "canceled"}
)
