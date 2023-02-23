package remindernlqparser

import (
	"context"
	"fmt"
	"regexp"
	"remindme/internal/core/domain/reminder"
	"strconv"
	"strings"
	"sync"
	"time"
)

var once sync.Once

var (
	reEvery  = regexp.MustCompile(`every(\s+(\d{1,3}))?\s*(\S+)($|\s+)`)
	reAtTime = regexp.MustCompile(
		strings.Join(
			[]string{
				`(noon|midday|midnight)`,
				`((\d{1,2})\s*(am|pm))`,
				`(at\s*(\d{1,2})\s*(am|pm)?)`,
				`((at)?\s*(\d{1,2}):(\d{2})\s*(am|pm)?)`,
			},
			"|",
		),
	)
	reOnTime = regexp.MustCompile(
		strings.Join(
			[]string{
				"today",
				"tomorrow|tmr|tmrw",
				"sunday|sun",
				"monday|mon",
				"tuesaday|tu|tue|tues",
				"wednesday|wed",
				"thursday|th|thu|thur|thurs",
				"friday|fri",
				"saturday|sat",
			},
			"|",
		),
	)
	reInTime = regexp.MustCompile(`(in|after)?\s*(\d{1,3})?\s*(\S+)($|\s+)`)
)

type Parser struct{}

func New() reminder.NaturalLanguageQueryParser {
	once.Do(func() {
		reEvery.Longest()
		reAtTime.Longest()
		reOnTime.Longest()
		reInTime.Longest()
	})
	return &Parser{}
}

func (p *Parser) Parse(
	ctx context.Context,
	query string,
	userLocalTime time.Time,
) (params reminder.CreateReminderParams, err error) {
	query = strings.ToLower(strings.TrimSpace(query))

	everyMatch := reEvery.FindStringSubmatchIndex(query)
	if everyMatch != nil {
		every, err := p.parseEvery(everyMatch, query)
		if err != nil {
			return params, err
		}

		atQuery := query[everyMatch[1]:]
		atMatch := reAtTime.FindStringSubmatchIndex(atQuery)
		if atMatch != nil {
			at, err := p.parseAt(atMatch, atQuery)
			if err != nil {
				return params, err
			}
			every.at = &at
		}

		return createParams(every, userLocalTime)
	}

	onMatch := reOnTime.FindStringSubmatchIndex(query)
	if onMatch != nil {
		on, err := p.parseOn(onMatch, query)
		if err != nil {
			return params, err
		}

		atQuery := query[onMatch[1]:]
		atMatch := reAtTime.FindStringSubmatchIndex(atQuery)
		if atMatch != nil {
			at, err := p.parseAt(atMatch, atQuery)
			if err != nil {
				return params, err
			}
			on.at = &at
		}

		return createParams(on, userLocalTime)
	}

	atMatch := reAtTime.FindStringSubmatchIndex(query)
	if atMatch != nil {
		at, err := p.parseAt(atMatch, query)
		if err != nil {
			return params, err
		}
		return createParams(at, userLocalTime)
	}

	inMatch := reInTime.FindStringSubmatchIndex(query)
	if inMatch != nil {
		in, err := p.parseIn(inMatch, query)
		if err != nil {
			return params, err
		}
		return createParams(in, userLocalTime)
	}

	return params, reminder.ErrNaturalQueryParsing
}

func (p *Parser) parseAt(match []int, query string) (at, error) {
	if len(match) != 26 {
		return at{}, fmt.Errorf("invalid match for parseAt, %w", reminder.ErrNaturalQueryParsing)
	}

	if match[2] != -1 {
		switch query[match[2]:match[3]] {
		case "midday":
			fallthrough
		case "noon":
			return at{hour: 12, minute: 0}, nil
		case "midnight":
			return at{}, nil
		}
		return at{}, reminder.ErrNaturalQueryParsing
	}

	if match[4] != -1 {
		rawHour := query[match[6]:match[7]]
		pmOrAm := query[match[8]:match[9]]
		return p.parseRawTimeAmOrPm(rawHour, "", pmOrAm)
	}

	if match[10] != -1 {
		rawHour := query[match[12]:match[13]]
		var pmOrAm string
		if match[14] != -1 {
			pmOrAm = query[match[14]:match[15]]
		}
		return p.parseRawTimeAmOrPm(rawHour, "", pmOrAm)
	}

	if match[16] != -1 {
		rawHour := query[match[20]:match[21]]
		rawMinute := query[match[22]:match[23]]
		var pmOrAm string
		if match[24] != -1 {
			pmOrAm = query[match[24]:match[25]]
		}
		return p.parseRawTimeAmOrPm(rawHour, rawMinute, pmOrAm)
	}

	return at{}, reminder.ErrNaturalQueryParsing
}

func (p *Parser) parseRawTimeAmOrPm(
	rawHour string,
	rawMinute string,
	pmOrAm string,
) (at, error) {
	atHour, err := strconv.ParseUint(rawHour, 10, 8)
	if err != nil {
		return at{}, reminder.ErrNaturalQueryParsing
	}

	var atMinute uint64
	if rawMinute != "" {
		atMinute, err = strconv.ParseUint(rawMinute, 10, 8)
		if err != nil {
			return at{}, reminder.ErrNaturalQueryParsing
		}
	}

	if pmOrAm != "" && (atHour < 1 || atHour > 12) {
		return at{}, reminder.ErrNaturalQueryParsing
	}
	if pmOrAm == "am" && atHour == 12 {
		atHour = 0
	}
	if pmOrAm == "pm" && atHour != 12 {
		atHour += 12
	}

	return at{hour: uint(atHour), minute: uint(atMinute)}, nil
}

func (p *Parser) parseEvery(match []int, query string) (every, error) {
	if len(match) != 10 {
		return every{}, fmt.Errorf("invalid match for parseEvery, %w", reminder.ErrNaturalQueryParsing)
	}

	var n uint = 1
	if match[4] != -1 {
		rawN := query[match[4]:match[5]]
		parsedN, err := strconv.ParseUint(rawN, 10, 16)
		if err != nil {
			return every{}, reminder.ErrNaturalQueryParsing
		}
		n = uint(parsedN)
	}

	var period period
	switch query[match[6]:match[7]] {
	case "m":
		fallthrough
	case "min":
		fallthrough
	case "mins":
		fallthrough
	case "minutes":
		fallthrough
	case "minute":
		period = minute
	case "h":
		fallthrough
	case "hours":
		fallthrough
	case "hour":
		period = hour
	case "d":
		fallthrough
	case "days":
		fallthrough
	case "day":
		period = day
	}
	if period == invalid {
		return every{}, fmt.Errorf("invalid every period, %w", reminder.ErrNaturalQueryParsing)
	}

	return every{n: n, p: period}, nil
}

func (p *Parser) parseOn(match []int, query string) (on, error) {
	if len(match) != 2 {
		return on{}, fmt.Errorf("invalid match for parseOn, %w", reminder.ErrNaturalQueryParsing)
	}

	switch query[match[0]:match[1]] {
	case "today":
		return on{day: today}, nil
	case "tmr":
		fallthrough
	case "tmrw":
		fallthrough
	case "tomorrow":
		return on{day: tomorrow}, nil
	case "sun":
		fallthrough
	case "sunday":
		return on{day: sunday}, nil
	case "mon":
		fallthrough
	case "monday":
		return on{day: monday}, nil
	case "tu":
		fallthrough
	case "tue":
		fallthrough
	case "tues":
		fallthrough
	case "tuesday":
		return on{day: tuesday}, nil
	case "wed":
		fallthrough
	case "wednesday":
		return on{day: wednesday}, nil
	case "th":
		fallthrough
	case "thu":
		fallthrough
	case "thur":
		fallthrough
	case "thurs":
		fallthrough
	case "thursday":
		return on{day: thursday}, nil
	case "fri":
		fallthrough
	case "friday":
		return on{day: friday}, nil
	case "sat":
		fallthrough
	case "saturday":
		return on{day: saturday}, nil
	default:
		return on{}, reminder.ErrNaturalQueryParsing
	}
}

func (p *Parser) parseIn(match []int, query string) (in, error) {
	if len(match) != 10 {
		return in{}, fmt.Errorf("invalid match for parseIn, %w", reminder.ErrNaturalQueryParsing)
	}

	var n uint = 1
	if match[4] != -1 {
		rawN := query[match[4]:match[5]]
		parsedN, err := strconv.ParseUint(rawN, 10, 16)
		if err != nil {
			return in{}, reminder.ErrNaturalQueryParsing
		}
		n = uint(parsedN)
	}

	var period period
	switch query[match[6]:match[7]] {
	case "d":
		fallthrough
	case "days":
		fallthrough
	case "day":
		period = day
	case "h":
		fallthrough
	case "hours":
		fallthrough
	case "hour":
		period = hour
	case "m":
		fallthrough
	case "min":
		fallthrough
	case "mins":
		fallthrough
	case "minutes":
		fallthrough
	case "minute":
		period = minute
	case "s":
		fallthrough
	case "sec":
		fallthrough
	case "secs":
		fallthrough
	case "seconds":
		fallthrough
	case "second":
		period = second
	}
	if period == invalid {
		return in{}, fmt.Errorf("invalid in period, %w", reminder.ErrNaturalQueryParsing)
	}

	return in{n: n, p: period}, nil
}
