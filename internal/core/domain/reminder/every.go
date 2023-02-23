package reminder

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/golang-module/carbon/v2"
)

var (
	ErrInvalidEvery = errors.New("invalid every")
	ErrParseEvery   = errors.New("invalid every string")
	ErrParsePeriod  = errors.New("invalid period")
)

type Period struct {
	v string
}

func ParsePeriod(value string) (p Period, err error) {
	switch value {
	case "minute":
		return PeriodMinute, nil
	case "hour":
		return PeriodHour, nil
	case "day":
		return PeriodDay, nil
	case "week":
		return PeriodWeek, nil
	case "month":
		return PeriodMonth, nil
	case "year":
		return PeriodYear, nil
	default:
		return PeriodUnknown, ErrParsePeriod
	}
}

var (
	PeriodUnknown Period = Period{}
	PeriodMinute  Period = Period{v: "minute"}
	PeriodHour    Period = Period{v: "hour"}
	PeriodDay     Period = Period{v: "day"}
	PeriodWeek    Period = Period{v: "week"}
	PeriodMonth   Period = Period{v: "month"}
	PeriodYear    Period = Period{v: "year"}
)

type Every struct {
	n      uint32
	period Period
}

func (e Every) String() string {
	return fmt.Sprintf("%d %s", e.n, e.period.v)
}

func (e Every) IsZero() bool {
	return e.n == 0 && e.period == PeriodUnknown
}

func ParseEvery(value string) (e Every, err error) {
	parts := strings.SplitN(value, " ", 2)
	if len(parts) != 2 {
		return e, ErrParseEvery
	}
	n, err := strconv.ParseUint(parts[0], 10, 32)
	if err != nil {
		return e, ErrParseEvery
	}
	period, err := ParsePeriod(parts[1])
	if err != nil {
		return e, err
	}
	e.n = uint32(n)
	e.period = period
	return e, nil
}

var (
	EveryMinute = Every{n: 1, period: PeriodMinute}
	EveryHour   = Every{n: 1, period: PeriodHour}
	EveryDay    = Every{n: 1, period: PeriodDay}
	EveryWeek   = Every{n: 1, period: PeriodWeek}
	EveryMonth  = Every{n: 1, period: PeriodMonth}
	EveryYear   = Every{n: 1, period: PeriodYear}
)

func NewEvery(n uint32, p Period) Every {
	return Every{n: n, period: p}
}

func (e Every) Validate() error {
	if e.n == 0 {
		return ErrInvalidEvery
	}
	if e.period == PeriodUnknown {
		return ErrInvalidEvery
	}
	totalDuration := e.TotalDuration()
	if totalDuration < EveryMinute.TotalDuration() || totalDuration > EveryYear.TotalDuration() {
		return ErrInvalidEvery
	}
	return nil
}

func (e Every) TotalDuration() time.Duration {
	switch e.period {
	case PeriodMinute:
		return time.Minute * time.Duration(e.n)
	case PeriodHour:
		return time.Hour * time.Duration(e.n)
	case PeriodDay:
		return 24 * time.Hour * time.Duration(e.n)
	case PeriodWeek:
		return 7 * 24 * time.Hour * time.Duration(e.n)
	case PeriodMonth:
		return 31 * 24 * time.Hour * time.Duration(e.n)
	case PeriodYear:
		return 372 * 24 * time.Hour * time.Duration(e.n)
	default:
		panic(fmt.Sprintf("unexpected period: %v", e.period))
	}
}

func (e Every) NextFrom(t time.Time) time.Time {
	switch e.period {
	case PeriodMinute, PeriodHour, PeriodDay:
		return t.Add(e.TotalDuration())
	case PeriodWeek:
		return carbon.Time2Carbon(t).AddWeeks(int(e.n)).Carbon2Time()
	case PeriodMonth:
		return carbon.Time2Carbon(t).AddMonthsNoOverflow(int(e.n)).Carbon2Time()
	case PeriodYear:
		return carbon.Time2Carbon(t).AddYearsNoOverflow(int(e.n)).Carbon2Time()
	default:
		panic(fmt.Sprintf("unexpected period: %v", e.period))
	}
}

func (e Every) PerDayCount() float64 {
	return float64(24*time.Hour) / float64(e.TotalDuration())
}
