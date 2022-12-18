package reminder

type Status struct {
	v string
}

var (
	Scheduled = Status{v: "scheduled"}
	Sent      = Status{v: "sent"}
	Canceled  = Status{v: "canceled"}
)
