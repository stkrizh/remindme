package channel

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

type FakeRepository struct {
	CreateReturnsError bool
	Created            []Channel
	ReadReturnsError   bool
	ReadChannels       []Channel
	GetByIDError       error
	GetByIDChannel     Channel
	CountReturnsError  bool
	CountChannels      uint
	Options            []ReadOptions
	UpdateError        error
	lock               sync.Mutex
}

func NewFakeRepository() *FakeRepository {
	return &FakeRepository{}
}

func (r *FakeRepository) Create(ctx context.Context, input CreateInput) (channel Channel, err error) {
	if r.CreateReturnsError {
		return channel, errors.New("coulf not create channel")
	}
	r.lock.Lock()
	defer r.lock.Unlock()
	channel = Channel{
		ID:                ID(1),
		Settings:          input.Settings,
		Type:              input.Type,
		CreatedBy:         input.CreatedBy,
		CreatedAt:         input.CreatedAt,
		VerificationToken: input.VerificationToken,
		VerifiedAt:        input.VerifiedAt,
	}
	r.Created = append(r.Created, channel)
	return channel, nil
}

func (r *FakeRepository) Read(ctx context.Context, options ReadOptions) (channels []Channel, err error) {
	if r.ReadReturnsError {
		return channels, fmt.Errorf("could not read channels")
	}
	r.lock.Lock()
	defer r.lock.Unlock()
	r.Options = append(r.Options, options)
	return r.ReadChannels, nil
}

func (r *FakeRepository) GetByID(ctx context.Context, id ID) (c Channel, err error) {
	if r.GetByIDError != nil {
		return c, r.GetByIDError
	}
	r.GetByIDChannel.ID = id
	return r.GetByIDChannel, nil
}

func (r *FakeRepository) Count(ctx context.Context, options ReadOptions) (count uint, err error) {
	if r.CountReturnsError {
		return count, fmt.Errorf("could not count channels")
	}
	r.lock.Lock()
	defer r.lock.Unlock()
	r.Options = append(r.Options, options)
	return r.CountChannels, nil
}

func (r *FakeRepository) Update(ctx context.Context, input UpdateInput) (channel Channel, err error) {
	if r.UpdateError != nil {
		return channel, r.UpdateError
	}
	channel.ID = input.ID
	if input.DoVerificationTokenUpdate {
		channel.VerificationToken = input.VerificationToken
	}
	if input.DoVerifiedAtUpdate {
		channel.VerifiedAt = input.VerifiedAt
	}
	if input.DoSettingsUpdate {
		channel.Settings = input.Settings
	}
	return channel, nil
}

type FakeVerificationTokenGenerator struct {
	Token VerificationToken
}

func NewFakeVerificationTokenGenerator(token VerificationToken) *FakeVerificationTokenGenerator {
	return &FakeVerificationTokenGenerator{Token: token}
}

func (g *FakeVerificationTokenGenerator) GenerateVerificationToken() VerificationToken {
	return g.Token
}

type FakeVerificationTokenSender struct {
	ReturnsError bool
	Sent         []VerificationToken
	SetChannels  []Channel
	lock         sync.Mutex
}

func NewFakeVerificationTokenSender() *FakeVerificationTokenSender {
	return &FakeVerificationTokenSender{}
}

func (g *FakeVerificationTokenSender) SendVerificationToken(
	ctx context.Context,
	token VerificationToken,
	channel Channel,
) error {
	if g.ReturnsError {
		return fmt.Errorf("could not send verification token")
	}
	g.lock.Lock()
	defer g.lock.Unlock()
	g.Sent = append(g.Sent, token)
	g.SetChannels = append(g.SetChannels, channel)
	return nil
}

type TestInternalChannelTokenGenerator struct {
	Token InternalChannelToken
}

func NewTestInternalChannelTokenGenerator(token string) *TestInternalChannelTokenGenerator {
	return &TestInternalChannelTokenGenerator{
		Token: InternalChannelToken(token),
	}
}

func (g *TestInternalChannelTokenGenerator) GenerateInternalChannelToken() InternalChannelToken {
	return g.Token
}
