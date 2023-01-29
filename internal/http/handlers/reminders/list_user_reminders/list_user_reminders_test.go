package listuserreminders

import (
	"context"
	"net/http"
	"net/http/httptest"
	"remindme/internal/core/domain/channel"
	c "remindme/internal/core/domain/common"
	"remindme/internal/core/domain/reminder"
	"remindme/internal/core/domain/user"
	service "remindme/internal/core/services/list_user_reminders"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var Reminders []reminder.ReminderWithChannels = []reminder.ReminderWithChannels{
	{
		Reminder: reminder.Reminder{
			ID:        reminder.ID(1),
			CreatedBy: user.ID(1),
			At:        time.Date(2020, 1, 2, 1, 1, 1, 0, time.UTC),
			CreatedAt: time.Date(2020, 1, 1, 1, 1, 1, 0, time.UTC),
			Status:    reminder.StatusCreated,
		},
		ChannelIDs: []channel.ID{100, 200},
	},
	{
		Reminder: reminder.Reminder{
			ID:        reminder.ID(2),
			CreatedBy: user.ID(1),
			At:        time.Date(2020, 1, 2, 1, 1, 1, 0, time.UTC),
			CreatedAt: time.Date(2020, 1, 1, 1, 1, 1, 0, time.UTC),
			Status:    reminder.StatusCreated,
		},
		ChannelIDs: []channel.ID{300},
	},
}

type stubService struct {
	reminders  []reminder.ReminderWithChannels
	totalCount uint
	err        error
	input      *service.Input
}

func newStubService() *stubService {
	return &stubService{
		reminders:  Reminders,
		totalCount: uint(len(Reminders)),
	}
}

func (s *stubService) Run(ctx context.Context, input service.Input) (result service.Result, err error) {
	if s.err != nil {
		return result, s.err
	}
	s.input = &input
	result.Reminders = s.reminders
	result.TotalCount = s.totalCount
	return result, nil
}

func TestListUserRemindersHandler(t *testing.T) {
	cases := []struct {
		url            string
		expectedStatus int
		expectedInput  *service.Input
	}{
		{
			url:            "/reminders",
			expectedStatus: http.StatusOK,
			expectedInput:  &service.Input{},
		},
		{
			url:            "/reminders?order_by=id_asc",
			expectedStatus: http.StatusOK,
			expectedInput:  &service.Input{OrderBy: reminder.OrderByIDAsc},
		},
		{
			url:            "/reminders?order_by=id_desc",
			expectedStatus: http.StatusOK,
			expectedInput:  &service.Input{OrderBy: reminder.OrderByIDDesc},
		},
		{
			url:            "/reminders?order_by=at_asc",
			expectedStatus: http.StatusOK,
			expectedInput:  &service.Input{OrderBy: reminder.OrderByAtAsc},
		},
		{
			url:            "/reminders?order_by=at_desc",
			expectedStatus: http.StatusOK,
			expectedInput:  &service.Input{OrderBy: reminder.OrderByAtDesc},
		},
		{
			url:            "/reminders?order_by=asd",
			expectedStatus: http.StatusBadRequest,
			expectedInput:  nil,
		},
		{
			url:            "/reminders?status_in=created",
			expectedStatus: http.StatusOK,
			expectedInput: &service.Input{
				StatusIn: c.NewOptional([]reminder.Status{reminder.StatusCreated}, true),
			},
		},
		{
			url:            "/reminders?status_in=scheduled",
			expectedStatus: http.StatusOK,
			expectedInput: &service.Input{
				StatusIn: c.NewOptional([]reminder.Status{reminder.StatusScheduled}, true),
			},
		},
		{
			url:            "/reminders?status_in=sent_success,canceled,sent_error,sent_limit_exceeded",
			expectedStatus: http.StatusOK,
			expectedInput: &service.Input{
				StatusIn: c.NewOptional([]reminder.Status{
					reminder.StatusSentSuccess,
					reminder.StatusCanceled,
					reminder.StatusSentError,
					reminder.StatusSentLimitExceeded,
				}, true),
			},
		},
		{
			url:            "/reminders?status_in=sending",
			expectedStatus: http.StatusOK,
			expectedInput: &service.Input{
				StatusIn: c.NewOptional([]reminder.Status{
					reminder.StatusSending,
				}, true),
			},
		},
		{
			url:            "/reminders?status_in=aaa",
			expectedStatus: http.StatusBadRequest,
			expectedInput:  nil,
		},
		{
			url:            "/reminders?status_in=created,scheduled,aaa",
			expectedStatus: http.StatusBadRequest,
			expectedInput:  nil,
		},
		{
			url:            "/reminders?limit=0",
			expectedStatus: http.StatusOK,
			expectedInput:  &service.Input{Limit: c.NewOptional[uint](0, true)},
		},
		{
			url:            "/reminders?limit=100",
			expectedStatus: http.StatusOK,
			expectedInput:  &service.Input{Limit: c.NewOptional[uint](100, true)},
		},
		{
			url:            "/reminders?limit=aaaa",
			expectedStatus: http.StatusBadRequest,
			expectedInput:  nil,
		},
		{
			url:            "/reminders?limit=101",
			expectedStatus: http.StatusBadRequest,
			expectedInput:  nil,
		},
		{
			url:            "/reminders?offset=0",
			expectedStatus: http.StatusOK,
			expectedInput:  &service.Input{Offset: 0},
		},
		{
			url:            "/reminders?offset=100",
			expectedStatus: http.StatusOK,
			expectedInput:  &service.Input{Offset: 100},
		},
		{
			url:            "/reminders?offset=asd",
			expectedStatus: http.StatusBadRequest,
			expectedInput:  nil,
		},
		{
			url:            "/reminders?offset=11122233344455534873458943758347598345",
			expectedStatus: http.StatusBadRequest,
			expectedInput:  nil,
		},
		{
			url:            "/reminders?status_in=created,scheduled&order_by=at_asc&limit=20&offset=40",
			expectedStatus: http.StatusOK,
			expectedInput: &service.Input{
				StatusIn: c.NewOptional([]reminder.Status{reminder.StatusCreated, reminder.StatusScheduled}, true),
				OrderBy:  reminder.OrderByAtAsc,
				Limit:    c.NewOptional[uint](20, true),
				Offset:   40,
			},
		},
	}

	for _, testcase := range cases {
		t.Run(testcase.url, func(t *testing.T) {
			req, err := http.NewRequest("GET", testcase.url, nil)
			if err != nil {
				t.Fatal(err)
			}

			service := newStubService()
			rr := httptest.NewRecorder()
			handler := New(service)
			handler.ServeHTTP(rr, req)

			assert.Equal(t, testcase.expectedStatus, rr.Code)
			assert.Equal(t, testcase.expectedInput, service.input)
		})
	}
}
