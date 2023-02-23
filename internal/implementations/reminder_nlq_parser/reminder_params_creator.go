package remindernlqparser

import (
	"fmt"
	c "remindme/internal/core/domain/common"
	"remindme/internal/core/domain/reminder"
	"time"

	"github.com/golang-module/carbon/v2"
)

func createParams(node node, userLocalTime time.Time) (params reminder.CreateReminderParams, err error) {
	creator := newReminderParamsCreator(carbon.Time2Carbon(userLocalTime))
	if err := node.accept(creator); err != nil {
		return params, err
	}
	params.At = creator.at.Carbon2Time()
	if !creator.every.IsZero() {
		params.Every = c.NewOptional(creator.every, true)
	}
	return params, nil
}

type reminderParamsCreator struct {
	userLocalTime carbon.Carbon
	at            carbon.Carbon
	every         reminder.Every
}

func newReminderParamsCreator(userLocalTime carbon.Carbon) *reminderParamsCreator {
	return &reminderParamsCreator{
		userLocalTime: userLocalTime,
		at:            userLocalTime,
	}
}

func (c *reminderParamsCreator) visitAt(at at) error {
	if at.hour > 24 {
		return fmt.Errorf("invalid at hour, %w", reminder.ErrNaturalQueryParsing)
	}
	if at.hour == 24 {
		at.hour = 0
	}
	if at.minute > 59 {
		return fmt.Errorf("invalid at minute, %w", reminder.ErrNaturalQueryParsing)
	}

	c.at = c.at.SetTimeMicro(int(at.hour), int(at.minute), 0, 0)
	if c.at.Lte(c.userLocalTime) {
		c.at = c.at.AddDay()
	}

	return nil
}

func (c *reminderParamsCreator) visitOn(on on) error {
	d := c.userLocalTime.DayOfWeek()
	if d == 0 {
		return fmt.Errorf("could not define user local day of week, %w", reminder.ErrNaturalQueryParsing)
	}

	var addDayCount int
	switch on.day {
	case today:
		// do nothing
	case tomorrow:
		addDayCount = 1
	case monday:
		addDayCount = 1 - d
	case tuesday:
		addDayCount = 2 - d
	case wednesday:
		addDayCount = 3 - d
	case thursday:
		addDayCount = 4 - d
	case friday:
		addDayCount = 5 - d
	case saturday:
		addDayCount = 6 - d
	case sunday:
		addDayCount = 7 - d
	default:
		return fmt.Errorf("on day is invalid, %w", reminder.ErrNaturalQueryParsing)
	}
	if addDayCount <= 0 {
		addDayCount += 7
	}

	c.at = c.at.AddDays(addDayCount)
	if on.at != nil {
		return on.at.accept(c)
	}

	return nil
}

func (c *reminderParamsCreator) visitEvery(every every) error {
	var e reminder.Every

	switch every.p {
	case minute:
		e = reminder.NewEvery(uint32(every.n), reminder.PeriodMinute)
	case hour:
		e = reminder.NewEvery(uint32(every.n), reminder.PeriodHour)
	case day:
		e = reminder.NewEvery(uint32(every.n), reminder.PeriodDay)
	case week:
		e = reminder.NewEvery(uint32(every.n), reminder.PeriodWeek)
	case month:
		e = reminder.NewEvery(uint32(every.n), reminder.PeriodMonth)
	default:
		return fmt.Errorf("every period is invalid, %w", reminder.ErrNaturalQueryParsing)
	}

	if e.Validate() != nil {
		return fmt.Errorf("every is invalid, %w", reminder.ErrNaturalQueryParsing)
	}

	c.every = e
	if every.at != nil {
		return every.at.accept(c)
	}

	c.at = carbon.Time2Carbon(e.NextFrom(c.userLocalTime.Carbon2Time()))
	return nil
}

func (c *reminderParamsCreator) visitIn(in in) error {
	switch in.p {
	case second:
		c.at = c.at.AddSeconds(int(in.n))
	case minute:
		c.at = c.at.AddMinutes(int(in.n))
	case hour:
		c.at = c.at.AddHours(int(in.n))
	case day:
		c.at = c.at.AddDays(int(in.n))
	case week:
		c.at = c.at.AddWeeks(int(in.n))
	case month:
		c.at = c.at.AddMonths(int(in.n))
	default:
		return fmt.Errorf("in period is invalid, %w", reminder.ErrNaturalQueryParsing)
	}

	return nil
}
