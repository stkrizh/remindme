package reminder

import (
	"context"
	"sync"
	"time"
)

type TestReminderRepository struct {
	CreateError          error
	CreatedCount         int
	Created              Reminder
	CreatedID            ID
	GetByIDError         error
	GetByIDReminder      ReminderWithChannels
	ReadError            error
	ReadReminders        []ReminderWithChannels
	ReadWith             []ReadOptions
	CountError           error
	CountResult          uint
	CountWith            []ReadOptions
	ReminderBeforeUpdate Reminder
	UpdateError          error
	LockError            error
	LockWith             []ID
	ScheduleWith         []ScheduleInput
	ScheduleResult       []Reminder
	ScheduleError        error
	DeleteWith           []ID
	DeleteError          error
	lock                 sync.Mutex
}

func NewTestReminderRepository() *TestReminderRepository {
	return &TestReminderRepository{}
}

func (r *TestReminderRepository) Create(ctx context.Context, input CreateInput) (rem Reminder, err error) {
	if r.CreateError != nil {
		return rem, r.CreateError
	}
	rem.ID = r.CreatedID
	rem.CreatedBy = input.CreatedBy
	rem.CreatedAt = input.CreatedAt
	rem.At = input.At
	rem.Every = input.Every
	rem.Status = input.Status
	rem.Body = input.Body
	rem.ScheduledAt = input.ScheduledAt
	rem.SentAt = input.SentAt
	rem.CanceledAt = input.CanceledAt

	r.lock.Lock()
	defer r.lock.Unlock()
	r.CreatedCount++
	r.Created = rem

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
	rem = r.GetByIDReminder
	rem.ID = id
	return rem, nil
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
	rem = r.ReminderBeforeUpdate
	rem.ID = input.ID
	if input.DoAtUpdate {
		rem.At = input.At
	}
	if input.DoEveryUpdate {
		rem.Every = input.Every
	}
	if input.DoBodyUpdate {
		rem.Body = input.Body
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

func (r *TestReminderRepository) Schedule(ctx context.Context, input ScheduleInput) ([]Reminder, error) {
	if r.ScheduleError != nil {
		return nil, r.ScheduleError
	}
	r.lock.Lock()
	defer r.lock.Unlock()
	r.ScheduleWith = append(r.ScheduleWith, input)
	return r.ScheduleResult, nil
}

func (r *TestReminderRepository) Delete(ctx context.Context, id ID) error {
	if r.DeleteError != nil {
		return r.DeleteError
	}
	r.lock.Lock()
	defer r.lock.Unlock()
	r.DeleteWith = append(r.DeleteWith, id)
	return nil
}

type TestReminderChannelRepository struct {
	CreateError             error
	CreatedForReminder      ID
	WasCreateCalled         bool
	DeleteByReminderIDError error
	DeletedByReminderID     ID
	WasDeleteCalled         bool
}

func NewTestReminderChannelRepository() *TestReminderChannelRepository {
	return &TestReminderChannelRepository{}
}

func (r *TestReminderChannelRepository) Create(ctx context.Context, input CreateChannelsInput) (ChannelIDs, error) {
	if r.CreateError != nil {
		return nil, r.CreateError
	}
	r.WasCreateCalled = true
	r.CreatedForReminder = input.ReminderID
	return input.ChannelIDs, nil
}

func (r *TestReminderChannelRepository) DeleteByReminderID(ctx context.Context, reminderID ID) error {
	if r.DeleteByReminderIDError != nil {
		return r.DeleteByReminderIDError
	}
	r.WasDeleteCalled = true
	r.DeletedByReminderID = reminderID
	return nil
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

type TestReminderSender struct {
	Sent      []ReminderWithChannels
	SentError error
	lock      sync.Mutex
}

func NewTestReminderSender() *TestReminderSender {
	return &TestReminderSender{}
}

func (s *TestReminderSender) SendReminder(ctx context.Context, reminder ReminderWithChannels) error {
	if s.SentError != nil {
		return s.SentError
	}
	s.lock.Lock()
	defer s.lock.Unlock()
	s.Sent = append(s.Sent, reminder)
	return nil
}

type TestNLQParser struct {
	Params     CreateReminderParams
	ParseError error
	CalledWith []struct {
		Query         string
		UserLocalTime time.Time
	}
	lock sync.Mutex
}

func NewTestNLQParser() *TestNLQParser {
	return &TestNLQParser{}
}

func (p *TestNLQParser) Parse(
	tx context.Context,
	query string,
	userLocalTime time.Time,
) (params CreateReminderParams, err error) {
	if p.ParseError != nil {
		return params, p.ParseError
	}
	p.lock.Lock()
	defer p.lock.Unlock()
	p.CalledWith = append(p.CalledWith, struct {
		Query         string
		UserLocalTime time.Time
	}{Query: query, UserLocalTime: userLocalTime})
	return p.Params, nil
}
