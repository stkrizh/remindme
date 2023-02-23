package remindernlqparser

import (
	"context"
	c "remindme/internal/core/domain/common"
	"remindme/internal/core/domain/reminder"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestCreateReminderParamsParsedSuccessfully(t *testing.T) {
	cases := []struct {
		query          string
		now            time.Time
		expectedParams reminder.CreateReminderParams
	}{
		{
			query: "at 20",
			now:   time.Date(2020, 1, 15, 15, 0, 0, 0, time.UTC),
			expectedParams: reminder.CreateReminderParams{
				At: time.Date(2020, 1, 15, 20, 0, 0, 0, time.UTC),
			},
		},
		{
			query: "8pm",
			now:   time.Date(2020, 1, 15, 15, 0, 0, 0, time.UTC),
			expectedParams: reminder.CreateReminderParams{
				At: time.Date(2020, 1, 15, 20, 0, 0, 0, time.UTC),
			},
		},
		{
			query: "at14",
			now:   time.Date(2020, 1, 31, 15, 33, 12, 10, tz("Europe/Kaliningrad")),
			expectedParams: reminder.CreateReminderParams{
				At: time.Date(2020, 2, 1, 14, 0, 0, 0, tz("Europe/Kaliningrad")),
			},
		},
		{
			query: "2 pm",
			now:   time.Date(2020, 1, 31, 15, 33, 12, 10, tz("Europe/Kaliningrad")),
			expectedParams: reminder.CreateReminderParams{
				At: time.Date(2020, 2, 1, 14, 0, 0, 0, tz("Europe/Kaliningrad")),
			},
		},
		{
			query: " at 3pm ",
			now:   time.Date(2020, 1, 31, 15, 33, 12, 10, tz("Europe/Kaliningrad")),
			expectedParams: reminder.CreateReminderParams{
				At: time.Date(2020, 2, 1, 15, 0, 0, 0, tz("Europe/Kaliningrad")),
			},
		},
		{
			query: "at 0",
			now:   time.Date(2020, 1, 31, 15, 0, 0, 0, time.UTC),
			expectedParams: reminder.CreateReminderParams{
				At: time.Date(2020, 2, 1, 0, 0, 0, 0, time.UTC),
			},
		},
		{
			query: "12am",
			now:   time.Date(2020, 1, 31, 15, 0, 0, 0, time.UTC),
			expectedParams: reminder.CreateReminderParams{
				At: time.Date(2020, 2, 1, 0, 0, 0, 0, time.UTC),
			},
		},
		{
			query: "12pm",
			now:   time.Date(2020, 1, 31, 15, 0, 0, 0, time.UTC),
			expectedParams: reminder.CreateReminderParams{
				At: time.Date(2020, 2, 1, 12, 0, 0, 0, time.UTC),
			},
		},
		{
			query: "   12:01 pm ",
			now:   time.Date(2020, 1, 31, 15, 0, 0, 0, time.UTC),
			expectedParams: reminder.CreateReminderParams{
				At: time.Date(2020, 2, 1, 12, 1, 0, 0, time.UTC),
			},
		},
		{
			query: "at 24",
			now:   time.Date(2020, 1, 31, 1, 45, 10, 20, tz("Europe/Kaliningrad")),
			expectedParams: reminder.CreateReminderParams{
				At: time.Date(2020, 2, 1, 0, 0, 0, 0, tz("Europe/Kaliningrad")),
			},
		},
		{
			query: " 20:00",
			now:   time.Date(2020, 1, 31, 19, 45, 10, 20, time.UTC),
			expectedParams: reminder.CreateReminderParams{
				At: time.Date(2020, 1, 31, 20, 0, 0, 0, time.UTC),
			},
		},
		{
			query: "at 20:33",
			now:   time.Date(2020, 1, 31, 19, 45, 10, 20, tz("Europe/Kaliningrad")),
			expectedParams: reminder.CreateReminderParams{
				At: time.Date(2020, 1, 31, 20, 33, 0, 0, tz("Europe/Kaliningrad")),
			},
		},
		{
			query: "07:46pm!",
			now:   time.Date(2020, 1, 31, 19, 45, 10, 20, tz("Europe/Kaliningrad")),
			expectedParams: reminder.CreateReminderParams{
				At: time.Date(2020, 1, 31, 19, 46, 0, 0, tz("Europe/Kaliningrad")),
			},
		},
		{
			query: "at 7:45  pm ",
			now:   time.Date(2020, 1, 31, 19, 45, 10, 20, tz("Europe/Kaliningrad")),
			expectedParams: reminder.CreateReminderParams{
				At: time.Date(2020, 2, 1, 19, 45, 0, 0, tz("Europe/Kaliningrad")),
			},
		},
		{
			query: " 19:46",
			now:   time.Date(2020, 1, 31, 19, 45, 10, 20, time.UTC),
			expectedParams: reminder.CreateReminderParams{
				At: time.Date(2020, 1, 31, 19, 46, 0, 0, time.UTC),
			},
		},
		{
			query: "at 00:00",
			now:   time.Date(2020, 1, 31, 10, 1, 2, 3, tz("Europe/Kaliningrad")),
			expectedParams: reminder.CreateReminderParams{
				At: time.Date(2020, 2, 1, 0, 0, 0, 0, tz("Europe/Kaliningrad")),
			},
		},
		{
			query: "23:59 ",
			now:   time.Date(2020, 1, 31, 10, 1, 2, 3, time.UTC),
			expectedParams: reminder.CreateReminderParams{
				At: time.Date(2020, 1, 31, 23, 59, 0, 0, time.UTC),
			},
		},
		{
			query: "at 0",
			now:   time.Date(2020, 1, 31, 10, 1, 2, 3, time.UTC),
			expectedParams: reminder.CreateReminderParams{
				At: time.Date(2020, 2, 1, 0, 0, 0, 0, time.UTC),
			},
		},
		{
			query: "at 1pm",
			now:   time.Date(2020, 1, 31, 10, 1, 2, 3, time.UTC),
			expectedParams: reminder.CreateReminderParams{
				At: time.Date(2020, 1, 31, 13, 0, 0, 0, time.UTC),
			},
		},
		{
			query: "at 09:00",
			now:   time.Date(2020, 1, 31, 10, 1, 2, 3, tz("Europe/Kaliningrad")),
			expectedParams: reminder.CreateReminderParams{
				At: time.Date(2020, 2, 1, 9, 0, 0, 0, tz("Europe/Kaliningrad")),
			},
		},
		{
			query: "at 9",
			now:   time.Date(2020, 1, 31, 10, 1, 2, 3, tz("Europe/Kaliningrad")),
			expectedParams: reminder.CreateReminderParams{
				At: time.Date(2020, 2, 1, 9, 0, 0, 0, tz("Europe/Kaliningrad")),
			},
		},
		{
			query: "at 9pm",
			now:   time.Date(2020, 1, 31, 10, 1, 2, 3, tz("Europe/Kaliningrad")),
			expectedParams: reminder.CreateReminderParams{
				At: time.Date(2020, 1, 31, 21, 0, 0, 0, tz("Europe/Kaliningrad")),
			},
		},
		{
			query: "  tmr ",
			now:   time.Date(2020, 1, 31, 10, 1, 2, 3, time.UTC),
			expectedParams: reminder.CreateReminderParams{
				At: time.Date(2020, 2, 1, 10, 1, 2, 3, time.UTC),
			},
		},
		{
			query: " tmrw! ",
			now:   time.Date(2020, 1, 31, 10, 1, 2, 3, time.UTC),
			expectedParams: reminder.CreateReminderParams{
				At: time.Date(2020, 2, 1, 10, 1, 2, 3, time.UTC),
			},
		},
		{
			query: "tomorrow",
			now:   time.Date(2020, 1, 31, 22, 3, 4, 5, time.UTC),
			expectedParams: reminder.CreateReminderParams{
				At: time.Date(2020, 2, 1, 22, 3, 4, 5, time.UTC),
			},
		},
		{
			query: "tomorrow 2pm",
			now:   time.Date(2020, 1, 31, 10, 1, 2, 3, tz("Europe/Kaliningrad")),
			expectedParams: reminder.CreateReminderParams{
				At: time.Date(2020, 2, 1, 14, 0, 0, 0, tz("Europe/Kaliningrad")),
			},
		},
		{
			query: "tmr midnight",
			now:   time.Date(2020, 1, 31, 10, 1, 2, 3, tz("Europe/Kaliningrad")),
			expectedParams: reminder.CreateReminderParams{
				At: time.Date(2020, 2, 1, 0, 0, 0, 0, tz("Europe/Kaliningrad")),
			},
		},
		{
			query: " tomorrow midday ",
			now:   time.Date(2020, 1, 31, 10, 1, 2, 3, tz("Europe/Kaliningrad")),
			expectedParams: reminder.CreateReminderParams{
				At: time.Date(2020, 2, 1, 12, 0, 0, 0, tz("Europe/Kaliningrad")),
			},
		},
		{
			query: " tomorrow at noon ",
			now:   time.Date(2020, 1, 31, 10, 1, 2, 3, tz("Europe/Kaliningrad")),
			expectedParams: reminder.CreateReminderParams{
				At: time.Date(2020, 2, 1, 12, 0, 0, 0, tz("Europe/Kaliningrad")),
			},
		},
		{
			query: "tomorrow at 0",
			now:   time.Date(2020, 1, 31, 10, 1, 2, 3, tz("Europe/Kaliningrad")),
			expectedParams: reminder.CreateReminderParams{
				At: time.Date(2020, 2, 1, 0, 0, 0, 0, tz("Europe/Kaliningrad")),
			},
		},
		{
			query: "tmr 1am",
			now:   time.Date(2020, 1, 31, 10, 1, 2, 3, tz("Europe/Kaliningrad")),
			expectedParams: reminder.CreateReminderParams{
				At: time.Date(2020, 2, 1, 1, 0, 0, 0, tz("Europe/Kaliningrad")),
			},
		},
		{
			query: "tomorrow 00:00",
			now:   time.Date(2020, 1, 31, 10, 1, 2, 3, tz("Europe/Kaliningrad")),
			expectedParams: reminder.CreateReminderParams{
				At: time.Date(2020, 2, 1, 0, 0, 0, 0, tz("Europe/Kaliningrad")),
			},
		},
		{
			query: "tmr 00:01",
			now:   time.Date(2020, 1, 31, 10, 1, 2, 3, tz("Europe/Kaliningrad")),
			expectedParams: reminder.CreateReminderParams{
				At: time.Date(2020, 2, 1, 0, 1, 0, 0, tz("Europe/Kaliningrad")),
			},
		},
		{
			query: "tomorrow 9am",
			now:   time.Date(2020, 1, 31, 10, 1, 2, 3, tz("Europe/Kaliningrad")),
			expectedParams: reminder.CreateReminderParams{
				At: time.Date(2020, 2, 1, 9, 0, 0, 0, tz("Europe/Kaliningrad")),
			},
		},
		{
			query: "tmr 15:33",
			now:   time.Date(2020, 1, 31, 10, 1, 2, 3, tz("Europe/Kaliningrad")),
			expectedParams: reminder.CreateReminderParams{
				At: time.Date(2020, 2, 1, 15, 33, 0, 0, tz("Europe/Kaliningrad")),
			},
		},
		{
			query: " tomorrow at 6",
			now:   time.Date(2020, 1, 31, 10, 1, 2, 3, tz("Europe/Kaliningrad")),
			expectedParams: reminder.CreateReminderParams{
				At: time.Date(2020, 2, 1, 6, 0, 0, 0, tz("Europe/Kaliningrad")),
			},
		},
		{
			query: " tomorrow at 6 pm ",
			now:   time.Date(2020, 1, 31, 10, 1, 2, 3, tz("Europe/Kaliningrad")),
			expectedParams: reminder.CreateReminderParams{
				At: time.Date(2020, 2, 1, 18, 0, 0, 0, tz("Europe/Kaliningrad")),
			},
		},
		{
			query: "mon",
			now:   time.Date(2023, 2, 28, 22, 3, 4, 5, time.UTC), // Tuesday
			expectedParams: reminder.CreateReminderParams{
				At: time.Date(2023, 3, 6, 22, 3, 4, 5, time.UTC),
			},
		},
		{
			query: "on monday ",
			now:   time.Date(2023, 2, 28, 11, 3, 4, 5, tz("Europe/Kaliningrad")), // Tuesday
			expectedParams: reminder.CreateReminderParams{
				At: time.Date(2023, 3, 6, 11, 3, 4, 5, tz("Europe/Kaliningrad")),
			},
		},
		{
			query: "on mon at 6:00 ",
			now:   time.Date(2023, 2, 28, 11, 3, 4, 5, tz("Europe/Kaliningrad")), // Tuesday
			expectedParams: reminder.CreateReminderParams{
				At: time.Date(2023, 3, 6, 6, 0, 0, 0, tz("Europe/Kaliningrad")),
			},
		},
		{
			query: "monday 3:30pm",
			now:   time.Date(2023, 2, 28, 11, 3, 4, 5, tz("Europe/Kaliningrad")), // Tuesday
			expectedParams: reminder.CreateReminderParams{
				At: time.Date(2023, 3, 6, 15, 30, 0, 0, tz("Europe/Kaliningrad")),
			},
		},
		{
			query: "tue",
			now:   time.Date(2023, 2, 28, 22, 3, 4, 5, time.UTC), // Tuesday
			expectedParams: reminder.CreateReminderParams{
				At: time.Date(2023, 3, 7, 22, 3, 4, 5, time.UTC),
			},
		},
		{
			query: "tuesday  22:22 ",
			now:   time.Date(2023, 2, 28, 22, 3, 4, 5, time.UTC), // Tuesday
			expectedParams: reminder.CreateReminderParams{
				At: time.Date(2023, 3, 7, 22, 22, 0, 0, time.UTC),
			},
		},
		{
			query: "on wed",
			now:   time.Date(2023, 2, 28, 22, 3, 4, 5, time.UTC), // Tuesday
			expectedParams: reminder.CreateReminderParams{
				At: time.Date(2023, 3, 1, 22, 3, 4, 5, time.UTC),
			},
		},
		{
			query: "on wed at 9",
			now:   time.Date(2023, 2, 28, 22, 3, 4, 5, time.UTC), // Tuesday
			expectedParams: reminder.CreateReminderParams{
				At: time.Date(2023, 3, 1, 9, 0, 0, 0, time.UTC),
			},
		},
		{
			query: "th",
			now:   time.Date(2023, 2, 28, 22, 3, 4, 5, time.UTC), // Tuesday
			expectedParams: reminder.CreateReminderParams{
				At: time.Date(2023, 3, 2, 22, 3, 4, 5, time.UTC),
			},
		},
		{
			query: "thur 11pm",
			now:   time.Date(2023, 2, 28, 22, 3, 4, 5, time.UTC), // Tuesday
			expectedParams: reminder.CreateReminderParams{
				At: time.Date(2023, 3, 2, 23, 0, 0, 0, time.UTC),
			},
		},
		{
			query: " onthursday at18:37",
			now:   time.Date(2023, 2, 28, 22, 3, 4, 5, time.UTC), // Tuesday
			expectedParams: reminder.CreateReminderParams{
				At: time.Date(2023, 3, 2, 18, 37, 0, 0, time.UTC),
			},
		},
		{
			query: "friday noon",
			now:   time.Date(2023, 2, 28, 22, 3, 4, 5, time.UTC), // Tuesday
			expectedParams: reminder.CreateReminderParams{
				At: time.Date(2023, 3, 3, 12, 0, 0, 0, time.UTC),
			},
		},
		{
			query: " fri at midnight",
			now:   time.Date(2023, 2, 28, 22, 3, 4, 5, time.UTC), // Tuesday
			expectedParams: reminder.CreateReminderParams{
				At: time.Date(2023, 3, 3, 0, 0, 0, 0, time.UTC),
			},
		},
		{
			query: " on friday midday ",
			now:   time.Date(2023, 2, 28, 22, 3, 4, 5, time.UTC), // Tuesday
			expectedParams: reminder.CreateReminderParams{
				At: time.Date(2023, 3, 3, 12, 0, 0, 0, time.UTC),
			},
		},
		{
			query: "sat 3:00 pm ",
			now:   time.Date(2023, 2, 28, 22, 3, 4, 5, time.UTC), // Tuesday
			expectedParams: reminder.CreateReminderParams{
				At: time.Date(2023, 3, 4, 15, 0, 0, 0, time.UTC),
			},
		},
		{
			query: "saturday 16:45:15",
			now:   time.Date(2023, 2, 28, 22, 3, 4, 5, time.UTC), // Tuesday
			expectedParams: reminder.CreateReminderParams{
				At: time.Date(2023, 3, 4, 16, 45, 0, 0, time.UTC),
			},
		},
		{
			query: " sunday 00:00",
			now:   time.Date(2023, 2, 28, 22, 3, 4, 5, time.UTC), // Tuesday
			expectedParams: reminder.CreateReminderParams{
				At: time.Date(2023, 3, 5, 0, 0, 0, 0, time.UTC),
			},
		},
		{
			query: " sun 12pm ",
			now:   time.Date(2023, 2, 28, 22, 3, 4, 5, time.UTC), // Tuesday
			expectedParams: reminder.CreateReminderParams{
				At: time.Date(2023, 3, 5, 12, 0, 0, 0, time.UTC),
			},
		},
		{
			query: " every day ",
			now:   time.Date(2023, 2, 28, 22, 3, 4, 5, time.UTC), // Tuesday
			expectedParams: reminder.CreateReminderParams{
				At:    time.Date(2023, 3, 1, 22, 3, 4, 5, time.UTC),
				Every: c.NewOptional(reminder.EveryDay, true),
			},
		},
		{
			query: " everyday at 3pm ",
			now:   time.Date(2023, 2, 28, 22, 3, 4, 5, time.UTC), // Tuesday
			expectedParams: reminder.CreateReminderParams{
				At:    time.Date(2023, 3, 1, 15, 0, 0, 0, time.UTC),
				Every: c.NewOptional(reminder.EveryDay, true),
			},
		},
		{
			query: " everyday at 12:01 pm ",
			now:   time.Date(2023, 2, 28, 11, 45, 4, 5, time.UTC), // Tuesday
			expectedParams: reminder.CreateReminderParams{
				At:    time.Date(2023, 2, 28, 12, 1, 0, 0, time.UTC),
				Every: c.NewOptional(reminder.EveryDay, true),
			},
		},
		{
			query: " every 3 days at 12:01 pm ",
			now:   time.Date(2023, 2, 28, 11, 45, 4, 5, time.UTC), // Tuesday
			expectedParams: reminder.CreateReminderParams{
				At:    time.Date(2023, 2, 28, 12, 1, 0, 0, time.UTC),
				Every: c.NewOptional(reminder.NewEvery(3, reminder.PeriodDay), true),
			},
		},
		{
			query: "  everyhour",
			now:   time.Date(2023, 2, 28, 11, 45, 4, 5, tz("Europe/Kaliningrad")), // Tuesday
			expectedParams: reminder.CreateReminderParams{
				At:    time.Date(2023, 2, 28, 12, 45, 4, 5, tz("Europe/Kaliningrad")),
				Every: c.NewOptional(reminder.EveryHour, true),
			},
		},
		{
			query: "  every 6 h",
			now:   time.Date(2023, 2, 28, 11, 45, 4, 5, tz("Europe/Kaliningrad")), // Tuesday
			expectedParams: reminder.CreateReminderParams{
				At:    time.Date(2023, 2, 28, 17, 45, 4, 5, tz("Europe/Kaliningrad")),
				Every: c.NewOptional(reminder.NewEvery(6, reminder.PeriodHour), true),
			},
		},
		{
			query: "every 48hours at 3pm",
			now:   time.Date(2023, 2, 28, 11, 45, 4, 5, tz("Europe/Kaliningrad")), // Tuesday
			expectedParams: reminder.CreateReminderParams{
				At:    time.Date(2023, 2, 28, 15, 0, 0, 0, tz("Europe/Kaliningrad")),
				Every: c.NewOptional(reminder.NewEvery(48, reminder.PeriodHour), true),
			},
		},
		{
			query: "every 3h 11:01",
			now:   time.Date(2023, 2, 28, 11, 45, 4, 5, tz("Europe/Kaliningrad")), // Tuesday
			expectedParams: reminder.CreateReminderParams{
				At:    time.Date(2023, 3, 1, 11, 1, 0, 0, tz("Europe/Kaliningrad")),
				Every: c.NewOptional(reminder.NewEvery(3, reminder.PeriodHour), true),
			},
		},
		{
			query: "every 7hour at noon",
			now:   time.Date(2023, 2, 28, 11, 45, 4, 5, tz("Europe/Kaliningrad")), // Tuesday
			expectedParams: reminder.CreateReminderParams{
				At:    time.Date(2023, 2, 28, 12, 0, 0, 0, tz("Europe/Kaliningrad")),
				Every: c.NewOptional(reminder.NewEvery(7, reminder.PeriodHour), true),
			},
		},
		{
			query: "everymin",
			now:   time.Date(2023, 2, 28, 11, 45, 4, 5, tz("Europe/Kaliningrad")), // Tuesday
			expectedParams: reminder.CreateReminderParams{
				At:    time.Date(2023, 2, 28, 11, 46, 4, 5, tz("Europe/Kaliningrad")),
				Every: c.NewOptional(reminder.NewEvery(1, reminder.PeriodMinute), true),
			},
		},
		{
			query: " every 10 min ",
			now:   time.Date(2023, 2, 28, 23, 55, 4, 5, tz("Europe/Kaliningrad")), // Tuesday
			expectedParams: reminder.CreateReminderParams{
				At:    time.Date(2023, 3, 1, 0, 5, 4, 5, tz("Europe/Kaliningrad")),
				Every: c.NewOptional(reminder.NewEvery(10, reminder.PeriodMinute), true),
			},
		},
		{
			query: " every 10 mins ",
			now:   time.Date(2023, 2, 28, 23, 55, 4, 5, tz("Europe/Kaliningrad")), // Tuesday
			expectedParams: reminder.CreateReminderParams{
				At:    time.Date(2023, 3, 1, 0, 5, 4, 5, tz("Europe/Kaliningrad")),
				Every: c.NewOptional(reminder.NewEvery(10, reminder.PeriodMinute), true),
			},
		},
		{
			query: " every 10 minutes ",
			now:   time.Date(2023, 2, 28, 23, 55, 4, 5, tz("Europe/Kaliningrad")), // Tuesday
			expectedParams: reminder.CreateReminderParams{
				At:    time.Date(2023, 3, 1, 0, 5, 4, 5, tz("Europe/Kaliningrad")),
				Every: c.NewOptional(reminder.NewEvery(10, reminder.PeriodMinute), true),
			},
		},
		{
			query: " every 15m at 3pm ",
			now:   time.Date(2023, 2, 28, 23, 55, 4, 5, tz("Europe/Kaliningrad")), // Tuesday
			expectedParams: reminder.CreateReminderParams{
				At:    time.Date(2023, 3, 1, 15, 0, 0, 0, tz("Europe/Kaliningrad")),
				Every: c.NewOptional(reminder.NewEvery(15, reminder.PeriodMinute), true),
			},
		},
		{
			query: " every 600m ",
			now:   time.Date(2023, 2, 28, 23, 55, 4, 5, tz("Europe/Kaliningrad")), // Tuesday
			expectedParams: reminder.CreateReminderParams{
				At:    time.Date(2023, 3, 1, 9, 55, 4, 5, tz("Europe/Kaliningrad")),
				Every: c.NewOptional(reminder.NewEvery(600, reminder.PeriodMinute), true),
			},
		},
		{
			query: " every 1m ",
			now:   time.Date(2023, 2, 28, 23, 55, 4, 5, tz("Europe/Kaliningrad")), // Tuesday
			expectedParams: reminder.CreateReminderParams{
				At:    time.Date(2023, 2, 28, 23, 56, 4, 5, tz("Europe/Kaliningrad")),
				Every: c.NewOptional(reminder.NewEvery(1, reminder.PeriodMinute), true),
			},
		},
		{
			query: "m",
			now:   time.Date(2023, 2, 28, 23, 55, 4, 5, tz("Europe/Kaliningrad")), // Tuesday
			expectedParams: reminder.CreateReminderParams{
				At: time.Date(2023, 2, 28, 23, 56, 4, 5, tz("Europe/Kaliningrad")),
			},
		},
		{
			query: "1m",
			now:   time.Date(2023, 2, 28, 23, 55, 4, 5, tz("Europe/Kaliningrad")), // Tuesday
			expectedParams: reminder.CreateReminderParams{
				At: time.Date(2023, 2, 28, 23, 56, 4, 5, tz("Europe/Kaliningrad")),
			},
		},
		{
			query: "in 10 m",
			now:   time.Date(2023, 2, 28, 23, 55, 4, 5, tz("Europe/Kaliningrad")), // Tuesday
			expectedParams: reminder.CreateReminderParams{
				At: time.Date(2023, 3, 1, 0, 5, 4, 5, tz("Europe/Kaliningrad")),
			},
		},
		{
			query: "2 min",
			now:   time.Date(2023, 2, 28, 23, 55, 4, 5, tz("Europe/Kaliningrad")), // Tuesday
			expectedParams: reminder.CreateReminderParams{
				At: time.Date(2023, 2, 28, 23, 57, 4, 5, tz("Europe/Kaliningrad")),
			},
		},
		{
			query: " in 5 mins",
			now:   time.Date(2023, 2, 28, 23, 55, 4, 5, tz("Europe/Kaliningrad")), // Tuesday
			expectedParams: reminder.CreateReminderParams{
				At: time.Date(2023, 3, 1, 0, 0, 4, 5, tz("Europe/Kaliningrad")),
			},
		},
		{
			query: " 5 minute  ",
			now:   time.Date(2023, 2, 28, 23, 55, 4, 5, tz("Europe/Kaliningrad")), // Tuesday
			expectedParams: reminder.CreateReminderParams{
				At: time.Date(2023, 3, 1, 0, 0, 4, 5, tz("Europe/Kaliningrad")),
			},
		},
		{
			query: " 24h  ",
			now:   time.Date(2023, 2, 28, 23, 55, 4, 5, time.UTC), // Tuesday
			expectedParams: reminder.CreateReminderParams{
				At: time.Date(2023, 3, 1, 23, 55, 4, 5, time.UTC),
			},
		},
		{
			query: " 2 hours  ",
			now:   time.Date(2023, 2, 28, 23, 55, 4, 5, time.UTC), // Tuesday
			expectedParams: reminder.CreateReminderParams{
				At: time.Date(2023, 3, 1, 1, 55, 4, 5, time.UTC),
			},
		},
		{
			query: "0h  ",
			now:   time.Date(2023, 2, 28, 23, 55, 4, 5, time.UTC), // Tuesday
			expectedParams: reminder.CreateReminderParams{
				At: time.Date(2023, 2, 28, 23, 55, 4, 5, time.UTC),
			},
		},
		{
			query: "in 48 hours  ",
			now:   time.Date(2023, 2, 28, 23, 55, 4, 5, time.UTC), // Tuesday
			expectedParams: reminder.CreateReminderParams{
				At: time.Date(2023, 3, 2, 23, 55, 4, 5, time.UTC),
			},
		},
		{
			query: "day",
			now:   time.Date(2023, 2, 28, 23, 55, 4, 5, time.UTC), // Tuesday
			expectedParams: reminder.CreateReminderParams{
				At: time.Date(2023, 3, 1, 23, 55, 4, 5, time.UTC),
			},
		},
		{
			query: "5 days",
			now:   time.Date(2023, 2, 28, 1, 2, 3, 4, time.UTC), // Tuesday
			expectedParams: reminder.CreateReminderParams{
				At: time.Date(2023, 3, 5, 1, 2, 3, 4, time.UTC),
			},
		},
		{
			query: "in 5d",
			now:   time.Date(2023, 2, 28, 1, 2, 3, 4, time.UTC), // Tuesday
			expectedParams: reminder.CreateReminderParams{
				At: time.Date(2023, 3, 5, 1, 2, 3, 4, time.UTC),
			},
		},
		{
			query: "after 2 days",
			now:   time.Date(2023, 2, 28, 1, 2, 3, 4, time.UTC), // Tuesday
			expectedParams: reminder.CreateReminderParams{
				At: time.Date(2023, 3, 2, 1, 2, 3, 4, time.UTC),
			},
		},
		{
			query: " 10s",
			now:   time.Date(2023, 2, 28, 1, 2, 3, 4, time.UTC), // Tuesday
			expectedParams: reminder.CreateReminderParams{
				At: time.Date(2023, 2, 28, 1, 2, 13, 4, time.UTC),
			},
		},
		{
			query: "in 300 secs",
			now:   time.Date(2023, 2, 28, 1, 2, 3, 4, time.UTC), // Tuesday
			expectedParams: reminder.CreateReminderParams{
				At: time.Date(2023, 2, 28, 1, 7, 3, 4, time.UTC),
			},
		},
		{
			query: "300 sec",
			now:   time.Date(2023, 2, 28, 1, 2, 3, 4, time.UTC), // Tuesday
			expectedParams: reminder.CreateReminderParams{
				At: time.Date(2023, 2, 28, 1, 7, 3, 4, time.UTC),
			},
		},
		{
			query: "60  second ",
			now:   time.Date(2023, 2, 28, 1, 2, 3, 4, time.UTC), // Tuesday
			expectedParams: reminder.CreateReminderParams{
				At: time.Date(2023, 2, 28, 1, 3, 3, 4, time.UTC),
			},
		},
		{
			query: " after 60 seconds",
			now:   time.Date(2023, 2, 28, 1, 2, 3, 4, time.UTC), // Tuesday
			expectedParams: reminder.CreateReminderParams{
				At: time.Date(2023, 2, 28, 1, 3, 3, 4, time.UTC),
			},
		},
		{
			query: "second",
			now:   time.Date(2023, 2, 28, 1, 2, 3, 4, time.UTC), // Tuesday
			expectedParams: reminder.CreateReminderParams{
				At: time.Date(2023, 2, 28, 1, 2, 4, 4, time.UTC),
			},
		},
	}

	parser := New()
	for _, testcase := range cases {
		id := testcase.query
		t.Run(id, func(t *testing.T) {
			params, err := parser.Parse(context.Background(), testcase.query, testcase.now)

			require.NoError(t, err)
			require.Equal(t, testcase.expectedParams, params)
		})
	}
}

func TestCreateReminderParamsParsingError(t *testing.T) {
	cases := []string{
		"",
		"sjdhfksd",
		" 48:00",
		"at 25:10",
		"23:60",
		"13am",
		" 13 pm",
		"13:99 pm",
		"every year",
		"everyyear",
		"every month",
		"everymonth",
		"every 10 month",
		"every 1 second",
		"every 999 days",
		"every 1234 hours",
		"every day at 25:55",
		"every 2 days at 10:60",
		"every 2 days at 13pm",
		"every 2 days at 13 am",
		"in 1234 hours",
		"mnday",
	}

	parser := New()
	for _, testcase := range cases {
		t.Run(testcase, func(t *testing.T) {
			_, err := parser.Parse(context.Background(), testcase, time.Now().UTC())
			require.ErrorIs(t, err, reminder.ErrNaturalQueryParsing)
		})
	}
}

func tz(name string) *time.Location {
	loc, err := time.LoadLocation(name)
	if err != nil {
		panic(err)
	}
	return loc
}
