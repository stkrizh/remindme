package reminder

import (
	"context"
	"sync"
)

type FakeReminderRepository struct {
	CreateError   error
	ReadError     error
	ReadReminders []Reminder
	ReadWith      []ReadOptions
	CountError    error
	CountResult   uint
	CountWith     []ReadOptions
	lock          sync.Mutex
}

func NewFakeReminderRepository() *FakeReminderRepository {
	return &FakeReminderRepository{}
}

func (r *FakeReminderRepository) Create(ctx context.Context, input CreateInput) (rem Reminder, err error) {
	if r.CreateError != nil {
		return rem, r.CreateError
	}
	rem.CreatedBy = input.CreatedBy
	rem.CreatedAt = input.CreatedAt
	rem.At = input.At
	rem.Every = input.Every
	rem.Status = input.Status
	return rem, err
}

func (r *FakeReminderRepository) Read(ctx context.Context, options ReadOptions) ([]Reminder, error) {
	if r.ReadError != nil {
		return nil, r.ReadError
	}
	r.lock.Lock()
	defer r.lock.Unlock()
	r.ReadWith = append(r.ReadWith, options)
	return r.ReadReminders, nil
}

func (r *FakeReminderRepository) Count(ctx context.Context, options ReadOptions) (uint, error) {
	if r.CountError != nil {
		return 0, r.CountError
	}
	r.lock.Lock()
	defer r.lock.Unlock()
	r.CountWith = append(r.CountWith, options)
	return r.CountResult, nil
}

type FakeReminderChannelRepository struct {
	CreateError        error
	CreatedForReminder ID
}

func NewFakeReminderChannelRepository() *FakeReminderChannelRepository {
	return &FakeReminderChannelRepository{}
}

func (r *FakeReminderChannelRepository) Create(ctx context.Context, input CreateChannelsInput) (ChannelIDs, error) {
	if r.CreateError != nil {
		return nil, r.CreateError
	}
	r.CreatedForReminder = input.ReminderID
	return input.ChannelIDs, nil
}
