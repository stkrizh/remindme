package reminder

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestEveryPredefinedVars(t *testing.T) {
	if EveryMinute.Validate() != nil {
		t.Fatal()
	}
	if EveryHour.Validate() != nil {
		t.Fatal()
	}
	if EveryDay.Validate() != nil {
		t.Fatal()
	}
	if EveryWeek.Validate() != nil {
		t.Fatal()
	}
	if EveryMonth.Validate() != nil {
		t.Fatal()
	}
	if EveryYear.Validate() != nil {
		t.Fatal()
	}
}

func TestEveryValid(t *testing.T) {
	cases := []struct {
		e Every
	}{
		{e: Every{n: 1, period: PeriodMinute}},
		{e: Every{n: 2, period: PeriodMinute}},
		{e: Every{n: 10, period: PeriodDay}},
		{e: Every{n: 72, period: PeriodHour}},
		{e: Every{n: 4, period: PeriodWeek}},
		{e: Every{n: 53, period: PeriodWeek}},
		{e: Every{n: 12, period: PeriodMonth}},
	}

	for _, testcase := range cases {
		t.Run(testcase.e.String(), func(t *testing.T) {
			err := testcase.e.Validate()
			if err != nil {
				t.Fatal(testcase.e, err)
			}
		})
	}
}

func TestEveryError(t *testing.T) {
	cases := []struct {
		e Every
	}{
		{e: Every{n: 1_000_000, period: PeriodMinute}},
		{e: Every{n: 0, period: PeriodMinute}},
		{e: Every{n: 0, period: PeriodHour}},
		{e: Every{n: 1_000_000, period: PeriodHour}},
		{e: Every{n: 0, period: PeriodDay}},
		{e: Every{n: 373, period: PeriodDay}},
		{e: Every{n: 0, period: PeriodWeek}},
		{e: Every{n: 54, period: PeriodWeek}},
		{e: Every{n: 0, period: PeriodWeek}},
		{e: Every{n: 13, period: PeriodMonth}},
		{e: Every{n: 0, period: PeriodYear}},
		{e: Every{n: 2, period: PeriodYear}},
	}

	for _, testcase := range cases {
		t.Run(testcase.e.String(), func(t *testing.T) {
			err := testcase.e.Validate()
			if err == nil {
				t.Fatal(testcase.e, err)
			}
		})
	}
}

func TestEveryNextFrom(t *testing.T) {
	cases := []struct {
		e        Every
		from     string
		expected string
	}{
		{
			e:        EveryMinute,
			from:     "2022-12-28T10:28:30+04:00",
			expected: "2022-12-28T10:29:30+04:00",
		},
		{
			e:        EveryHour,
			from:     "2022-12-31T23:28:30+00:00",
			expected: "2023-01-01T00:28:30+00:00",
		},
		{
			e:        EveryDay,
			from:     "2022-02-28T10:28:30+04:00",
			expected: "2022-03-01T10:28:30+04:00",
		},
		{
			e:        EveryWeek,
			from:     "2022-11-24T15:33:33+04:00",
			expected: "2022-12-01T15:33:33+04:00",
		},
		{
			e:        EveryWeek,
			from:     "2023-02-23T15:33:33+04:00",
			expected: "2023-03-02T15:33:33+04:00",
		},
		{
			e:        EveryMonth,
			from:     "2022-01-31T15:33:33+01:00",
			expected: "2022-02-28T15:33:33+01:00",
		},
		{
			e:        EveryMonth,
			from:     "2022-03-31T15:33:33-01:00",
			expected: "2022-04-30T15:33:33-01:00",
		},
		{
			e:        EveryYear,
			from:     "2022-01-31T00:00:00Z",
			expected: "2023-01-31T00:00:00Z",
		},
		{
			e:        EveryYear,
			from:     "2020-02-29T01:02:03+04:00",
			expected: "2021-02-28T01:02:03+04:00",
		},
		{
			e:        Every{n: 15, period: PeriodMinute},
			from:     "2022-12-31T23:50:03+04:00",
			expected: "2023-01-01T00:05:03+04:00",
		},
		{
			e:        Every{n: 92, period: PeriodMinute},
			from:     "2022-12-31T00:00:03+04:00",
			expected: "2022-12-31T01:32:03+04:00",
		},
		{
			e:        Every{n: 24, period: PeriodHour},
			from:     "2022-12-31T23:50:03-04:00",
			expected: "2023-01-01T23:50:03-04:00",
		},
		{
			e:        Every{n: 25, period: PeriodHour},
			from:     "2022-12-31T23:50:03-04:00",
			expected: "2023-01-02T00:50:03-04:00",
		},
		{
			e:        Every{n: 72, period: PeriodHour},
			from:     "2022-12-01T01:50:03+00:00",
			expected: "2022-12-04T01:50:03+00:00",
		},
		{
			e:        Every{n: 2, period: PeriodDay},
			from:     "2022-12-01T01:50:03+00:00",
			expected: "2022-12-03T01:50:03+00:00",
		},
		{
			e:        Every{n: 10, period: PeriodDay},
			from:     "2022-12-25T01:50:03+03:00",
			expected: "2023-01-04T01:50:03+03:00",
		},
		{
			e:        Every{n: 2, period: PeriodWeek},
			from:     "2023-02-17T01:02:03+04:00",
			expected: "2023-03-03T01:02:03+04:00",
		},
		{
			e:        Every{n: 4, period: PeriodWeek},
			from:     "2023-01-31T01:02:03+04:00",
			expected: "2023-02-28T01:02:03+04:00",
		},
		{
			e:        Every{n: 3, period: PeriodMonth},
			from:     "2022-11-30T01:02:03+04:00",
			expected: "2023-02-28T01:02:03+04:00",
		},
		{
			e:        Every{n: 4, period: PeriodMonth},
			from:     "2019-10-31T01:02:03+04:00",
			expected: "2020-02-29T01:02:03+04:00",
		},
		{
			e:        Every{n: 11, period: PeriodMonth},
			from:     "2020-02-29T01:02:03-04:00",
			expected: "2021-01-29T01:02:03-04:00",
		},
		{
			e:        Every{n: 12, period: PeriodMonth},
			from:     "2020-02-29T01:02:03+04:00",
			expected: "2021-02-28T01:02:03+04:00",
		},
	}

	for _, testcase := range cases {
		t.Run(testcase.e.String()+" "+testcase.from, func(t *testing.T) {
			from, err := time.Parse(time.RFC3339, testcase.from)
			assert.Nil(t, err, "invalid time format (testcase.from): %s", testcase.from)
			expected, err := time.Parse(time.RFC3339, testcase.expected)
			assert.Nil(t, err, "invalid time format (testcase.expected): %s", testcase.expected)
			assert.Equal(t, expected, testcase.e.NextFrom(from))
		})
	}
}

func TestParsePeriod(t *testing.T) {
	validCases := []struct {
		value    string
		expected Period
	}{
		{"minute", PeriodMinute},
		{"hour", PeriodHour},
		{"day", PeriodDay},
		{"week", PeriodWeek},
		{"month", PeriodMonth},
		{"year", PeriodYear},
	}
	for _, testcase := range validCases {
		t.Run(testcase.value, func(t *testing.T) {
			p, err := ParsePeriod(testcase.value)
			assert.Nil(t, err)
			assert.Equal(t, testcase.expected, p)
		})
	}

	invalidCases := []struct {
		value    string
		expected Period
	}{
		{"", PeriodUnknown},
		{" ", PeriodUnknown},
		{"day ", PeriodUnknown},
		{" week", PeriodUnknown},
		{"Month", PeriodUnknown},
		{"ear", PeriodUnknown},
	}
	for _, testcase := range invalidCases {
		t.Run(testcase.value, func(t *testing.T) {
			p, err := ParsePeriod(testcase.value)
			assert.ErrorIs(t, err, ErrParsePeriod)
			assert.Equal(t, testcase.expected, p)
		})
	}
}

func TestParseEverySuccess(t *testing.T) {
	cases := []struct {
		value    string
		expected Every
	}{
		{"1 minute", Every{1, PeriodMinute}},
		{"2 minute", Every{2, PeriodMinute}},
		{"100500 minute", Every{100500, PeriodMinute}},
		{"1 hour", Every{1, PeriodHour}},
		{"24 hour", Every{24, PeriodHour}},
		{"1000 hour", Every{1000, PeriodHour}},
		{"1 day", Every{1, PeriodDay}},
		{"7 day", Every{7, PeriodDay}},
		{"372 day", Every{372, PeriodDay}},
		{"1 week", Every{1, PeriodWeek}},
		{"4 week", Every{4, PeriodWeek}},
		{"53 week", Every{53, PeriodWeek}},
		{"1 month", Every{1, PeriodMonth}},
		{"6 month", Every{6, PeriodMonth}},
		{"12 month", Every{12, PeriodMonth}},
		{"1 year", Every{1, PeriodYear}},
	}

	for _, testcase := range cases {
		t.Run(testcase.value, func(t *testing.T) {
			every, err := ParseEvery(testcase.value)
			assert.Nil(t, err)
			assert.Equal(t, testcase.expected, every)
		})
	}
}

func TestEveryToStingAndParse(t *testing.T) {
	cases := []Every{
		EveryMinute,
		EveryHour,
		EveryDay,
		EveryWeek,
		EveryMonth,
		EveryYear,
		{2, PeriodMinute},
		{200_000, PeriodMinute},
		{30, PeriodHour},
		{1_123, PeriodHour},
		{300, PeriodDay},
		{2, PeriodWeek},
		{10, PeriodMonth},
		{1, PeriodYear},
	}

	for _, every := range cases {
		t.Run(every.String(), func(t *testing.T) {
			parsed, err := ParseEvery(every.String())
			assert.Nil(t, err)
			assert.Equal(t, every, parsed)
		})
	}
}

func TestParseEveryError(t *testing.T) {
	cases := []struct {
		value string
		err   error
	}{
		{"", ErrParseEvery},
		{" ", ErrParseEvery},
		{"1", ErrParseEvery},
		{"1 ", ErrParsePeriod},
		{"hour", ErrParseEvery},
		{"1minute", ErrParseEvery},
		{"1-minute", ErrParseEvery},
		{"1 minutes", ErrParsePeriod},
		{"1 hours", ErrParsePeriod},
		{"1 days", ErrParsePeriod},
		{"1 weeks", ErrParsePeriod},
		{"1 months", ErrParsePeriod},
		{"1 years", ErrParsePeriod},
		{"-1 minute", ErrParseEvery},
		{"5000000000 minute", ErrParseEvery},
	}

	for _, testcase := range cases {
		t.Run(testcase.value, func(t *testing.T) {
			_, err := ParseEvery(testcase.value)
			assert.ErrorIs(t, err, testcase.err)
		})
	}
}

func TestPerDayCount(t *testing.T) {
	cases := []struct {
		e        Every
		expected float64
	}{
		{NewEvery(5, PeriodMinute), 288},
		{NewEvery(20, PeriodMinute), 72},
		{NewEvery(60, PeriodMinute), 24},
		{NewEvery(1, PeriodHour), 24},
		{NewEvery(3, PeriodHour), 8},
		{NewEvery(24, PeriodHour), 1},
		{NewEvery(1, PeriodDay), 1},
		{NewEvery(10, PeriodDay), 0.1},
		{NewEvery(100, PeriodDay), 0.01},
		{NewEvery(1, PeriodWeek), 1.0 / 7},
		{NewEvery(1, PeriodMonth), 1.0 / 31},
		{NewEvery(1, PeriodYear), 1.0 / 372},
	}

	for _, testcase := range cases {
		t.Run(testcase.e.String(), func(t *testing.T) {
			assert.InDelta(t, testcase.expected, testcase.e.PerDayCount(), 0.001)
		})
	}
}
