package reminder

import (
	"context"
	"sync"
)

type TestReminderRepository struct {
	CreateError     error
	GetByIDError    error
	GetByIDReminder ReminderWithChannels
	ReadError       error
	ReadReminders   []ReminderWithChannels
	ReadWith        []ReadOptions
	CountError      error
	CountResult     uint
	CountWith       []ReadOptions
	UpdateError     error
	LockError       error
	LockWith        []ID
	lock            sync.Mutex
}

func NewTestReminderRepository() *TestReminderRepository {
	return &TestReminderRepository{}
}

func (r *TestReminderRepository) Create(ctx context.Context, input CreateInput) (rem Reminder, err error) {
	if r.CreateError != nil {
		return rem, r.CreateError
	}
	rem.CreatedBy = input.CreatedBy
	rem.CreatedAt = input.CreatedAt
	rem.At = input.At
	rem.Every = input.Every
	rem.Status = input.Status
	rem.ScheduledAt = input.ScheduledAt
	rem.SentAt = input.SentAt
	rem.CanceledAt = input.CanceledAt
	return rem, err
}

func (r *TestReminderRepository) Lock(ctx context.Context, id ID) error {
	if r.LockError != nil {
		return r.LockError
	}
	r.lock.Lock()
	defer r.lock.Unlock()

	r.LockWith = append(r.LockWith, id)
	return nil
}

func (r *TestReminderRepository) GetByID(ctx context.Context, id ID) (rem ReminderWithChannels, err error) {
	if r.GetByIDError != nil {
		return rem, r.GetByIDError
	}
	return r.GetByIDReminder, nil
}

func (r *TestReminderRepository) Read(ctx context.Context, options ReadOptions) ([]ReminderWithChannels, error) {
	if r.ReadError != nil {
		return nil, r.ReadError
	}
	r.lock.Lock()
	defer r.lock.Unlock()
	r.ReadWith = append(r.ReadWith, options)
	return r.ReadReminders, nil
}

func (r *TestReminderRepository) Count(ctx context.Context, options ReadOptions) (uint, error) {
	if r.CountError != nil {
		return 0, r.CountError
	}
	r.lock.Lock()
	defer r.lock.Unlock()
	r.CountWith = append(r.CountWith, options)
	return r.CountResult, nil
}

func (r *TestReminderRepository) Update(ctx context.Context, input UpdateInput) (rem Reminder, err error) {
	if r.UpdateError != nil {
		return rem, r.UpdateError
	}
	rem.ID = input.ID
	if input.DoAtUpdate {
		rem.At = input.At
	}
	if input.DoEveryUpdate {
		rem.Every = input.Every
	}
	if input.DoStatusUpdate {
		rem.Status = input.Status
	}
	if input.DoScheduledAtUpdate {
		rem.ScheduledAt = input.ScheduledAt
	}
	if input.DoSentAtUpdate {
		rem.SentAt = input.SentAt
	}
	if input.DoCanceledAtUpdate {
		rem.CanceledAt = input.CanceledAt
	}
	return rem, nil
}

type TestReminderChannelRepository struct {
	CreateError        error
	CreatedForReminder ID
}

func NewTestReminderChannelRepository() *TestReminderChannelRepository {
	return &TestReminderChannelRepository{}
}

func (r *TestReminderChannelRepository) Create(ctx context.Context, input CreateChannelsInput) (ChannelIDs, error) {
	if r.CreateError != nil {
		return nil, r.CreateError
	}
	r.CreatedForReminder = input.ReminderID
	return input.ChannelIDs, nil
}

type TestReminderScheduler struct {
	Scheduled []Reminder
	Error     error
	lock      sync.Mutex
}

func (s *TestReminderScheduler) ScheduleReminder(ctx context.Context, r Reminder) error {
	if s.Error != nil {
		return s.Error
	}
	s.lock.Lock()
	defer s.lock.Unlock()
	s.Scheduled = append(s.Scheduled, r)
	return nil
}

func NewTestReminderScheduler() *TestReminderScheduler {
	return &TestReminderScheduler{}
}
