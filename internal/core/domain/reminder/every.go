package reminder

import "time"

type Every struct{ d time.Duration }

var (
	EveryDay   = Every{d: time.Hour * 24}
	EveryWeek  = Every{d: time.Hour * 24 * 7}
	EveryMonth = Every{d: time.Hour * 24 * 31}
	EveryYear  = Every{d: time.Hour * 24 * 366}
)

func NewEvery(d time.Duration) (e Every, ok bool) {
	e = Every{d: d}
	if !e.IsValid() {
		return e, false
	}
	return e, true
}

func (e Every) IsValid() bool {
	return e.d >= time.Minute && e.d < time.Hour*24*366
}
