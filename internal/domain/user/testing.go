package user

import (
	"context"
	"crypto/md5"
	"fmt"
	"io"
	"sync"
)

type FakeActivationTokenSender struct {
	Sent        []User
	ReturnError bool
	lock        sync.RWMutex
}

func NewFakeActivationTokenSender() *FakeActivationTokenSender {
	return &FakeActivationTokenSender{}
}

func (s *FakeActivationTokenSender) SendToken(ctx context.Context, user User) error {
	if s.ReturnError {
		return fmt.Errorf("Could not send activation token for user %v", user)
	}
	s.lock.Lock()
	defer s.lock.Unlock()
	s.Sent = append(s.Sent, user)
	return nil
}

func (s *FakeActivationTokenSender) SentCount() int {
	return len(s.Sent)
}

func (s *FakeActivationTokenSender) LastSentTo() User {
	l := len(s.Sent)
	if l == 0 {
		panic("Sent count is 0.")
	}
	return s.Sent[l-1]
}

type FakeActivationTokenGenerator struct {
	Token ActivationToken
}

func NewFakeActivationTokenGenerator(token string) *FakeActivationTokenGenerator {
	return &FakeActivationTokenGenerator{Token: ActivationToken(token)}
}

func (g *FakeActivationTokenGenerator) GenerateToken() ActivationToken {
	return g.Token
}

type FakePasswordHasher struct{}

func NewFakePasswordHasher() *FakePasswordHasher {
	return &FakePasswordHasher{}
}

func (h *FakePasswordHasher) HashPassword(password RawPassword) PasswordHash {
	hash := md5.New()
	io.WriteString(hash, string(password))
	return PasswordHash(fmt.Sprintf("%x", hash.Sum(nil)))
}

func (h *FakePasswordHasher) ValidatePassword(password RawPassword, hash PasswordHash) bool {
	return h.HashPassword(password) == hash
}

type FakeRepository struct {
	Users       map[ID]User
	ReturnError bool
	lock        sync.RWMutex
}

func NewFakeRepository() *FakeRepository {
	return &FakeRepository{Users: make(map[ID]User)}
}

func (r *FakeRepository) Create(ctx context.Context, input CreateUserInput) (*User, error) {
	if r.ReturnError {
		return nil, fmt.Errorf("could not create user %v", input)
	}
	r.lock.Lock()
	defer r.lock.Unlock()
	maxID := ID(0)
	for id, u := range r.Users {
		if input.Email.IsPresent && u.Email == input.Email {
			return nil, &EmailAlreadyExistsError{Email: u.Email.Value}
		}
		if input.Identity.IsPresent && u.Identity == input.Identity {
			return nil, &IdentityAlreadyExistsError{Identity: u.Identity.Value}
		}
		if input.ActivationToken.IsPresent && u.ActivationToken == input.ActivationToken {
			return nil, &ActivationTokenAlreadyExistsError{ActivationToken: u.ActivationToken.Value}
		}
		maxID = id
	}
	user := User{
		ID:              maxID + 1,
		Email:           input.Email,
		PasswordHash:    input.PasswordHash,
		Identity:        input.Identity,
		CreatedAt:       input.CreatedAt,
		ActivatedAt:     input.ActivatedAt,
		ActivationToken: input.ActivationToken,
	}
	r.Users[user.ID] = user
	return &user, nil
}

func (r *FakeRepository) GetByID(ctx context.Context, id ID) (*User, error) {
	user, ok := r.Users[id]
	if !ok {
		return nil, &UserDoesNotExistError{}
	}
	return &user, nil
}
