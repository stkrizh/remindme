package getlimitforactivereminders

import (
	"context"
	c "remindme/internal/core/domain/common"
	"remindme/internal/core/domain/logging"
	"remindme/internal/core/domain/reminder"
	"remindme/internal/core/domain/user"
	"remindme/internal/core/services"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

var (
	NOW     = time.Date(2022, 6, 15, 12, 34, 55, 1, time.UTC)
	USER_ID = user.ID(1)
)

type Fixture struct {
	log       *logging.FakeLogger
	limits    *user.FakeLimitsRepository
	reminders *reminder.TestReminderRepository
}

func NewFixture() Fixture {
	return Fixture{
		log:       logging.NewFakeLogger(),
		limits:    user.NewFakeLimitsRepository(),
		reminders: reminder.NewTestReminderRepository(),
	}
}

func (f *Fixture) service() services.Service[Input, Result] {
	return New(f.log, f.limits, f.reminders)
}

func TestGetLimitForActiveRemindersSuccess(t *testing.T) {
	cases := []struct {
		id             string
		limits         user.Limits
		reminderCount  uint32
		expectedResult Result
	}{
		{
			id:             "1",
			limits:         user.Limits{},
			reminderCount:  100,
			expectedResult: Result{},
		},
		{
			id:             "2",
			limits:         user.Limits{ActiveReminderCount: c.NewOptional(uint32(200), true)},
			reminderCount:  100,
			expectedResult: Result{Limit: c.NewOptional(user.Limit{Value: 200, Actual: 100}, true)},
		},
		{
			id:             "3",
			limits:         user.Limits{ActiveReminderCount: c.NewOptional(uint32(50), true)},
			reminderCount:  200,
			expectedResult: Result{Limit: c.NewOptional(user.Limit{Value: 50, Actual: 200}, true)},
		},
	}

	for _, testcase := range cases {
		t.Run(testcase.id, func(t *testing.T) {
			fixture := NewFixture()
			fixture.limits.Limits = testcase.limits
			fixture.reminders.CountResult = uint(testcase.reminderCount)

			result, err := fixture.service().Run(context.Background(), Input{UserID: USER_ID})

			assert := require.New(t)
			assert.Nil(err)
			assert.Equal(testcase.expectedResult, result)
		})
	}
}

func TestRemindersCounted(t *testing.T) {
	fixture := NewFixture()
	fixture.limits.Limits = user.Limits{
		ActiveReminderCount:      c.NewOptional(uint32(5), true),
		MonthlySentReminderCount: c.NewOptional(uint32(10), true),
	}

	_, err := fixture.service().Run(context.Background(), Input{UserID: USER_ID})

	assert := require.New(t)
	assert.Nil(err)
	assert.Equal(
		[]reminder.ReadOptions{
			{
				CreatedByEquals: c.NewOptional(USER_ID, true),
				StatusIn: c.NewOptional(
					[]reminder.Status{reminder.StatusCreated, reminder.StatusScheduled},
					true,
				),
			},
		},
		fixture.reminders.CountWith,
	)
}
