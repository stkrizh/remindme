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
		return fmt.Errorf("could not send activation token for user %v", user)
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

func (h *FakePasswordHasher) HashPassword(password RawPassword) (PasswordHash, error) {
	hash := md5.New()
	io.WriteString(hash, string(password))
	return PasswordHash(fmt.Sprintf("%x", hash.Sum(nil))), nil
}

func (h *FakePasswordHasher) ValidatePassword(password RawPassword, hash PasswordHash) bool {
	actualHash, err := h.HashPassword(password)
	if err != nil {
		return false
	}
	return actualHash == hash
}

type FakeUserRepository struct {
	Users       map[ID]User
	ReturnError bool
	lock        sync.RWMutex
}

func NewFakeUserRepository() *FakeUserRepository {
	return &FakeUserRepository{Users: make(map[ID]User)}
}

func (r *FakeUserRepository) Create(ctx context.Context, input CreateUserInput) (u User, err error) {
	if r.ReturnError {
		return u, fmt.Errorf("could not create user %v", input)
	}
	r.lock.Lock()
	defer r.lock.Unlock()
	maxID := ID(0)
	for id, u := range r.Users {
		if input.Email.IsPresent && u.Email == input.Email {
			return u, &EmailAlreadyExistsError{Email: u.Email.Value}
		}
		if input.Identity.IsPresent && u.Identity == input.Identity {
			return u, &IdentityAlreadyExistsError{Identity: u.Identity.Value}
		}
		if input.ActivationToken.IsPresent && u.ActivationToken == input.ActivationToken {
			return u, &ActivationTokenAlreadyExistsError{ActivationToken: u.ActivationToken.Value}
		}
		maxID = id
	}
	u = User{
		ID:              maxID + 1,
		Email:           input.Email,
		PasswordHash:    input.PasswordHash,
		Identity:        input.Identity,
		CreatedAt:       input.CreatedAt,
		ActivatedAt:     input.ActivatedAt,
		ActivationToken: input.ActivationToken,
	}
	r.Users[u.ID] = u
	return u, nil
}

func (r *FakeUserRepository) GetByID(ctx context.Context, id ID) (u User, err error) {
	u, ok := r.Users[id]
	if !ok {
		return u, &UserDoesNotExistError{}
	}
	return u, nil
}

type FakeSessionRepository struct {
	UserIdByToken  map[SessionToken]ID
	UserRepository UserRepository
	ReturnError    bool
	lock           sync.RWMutex
}

func NewFakeSessionRepository(userRepository UserRepository) *FakeSessionRepository {
	return &FakeSessionRepository{
		UserIdByToken:  make(map[SessionToken]ID),
		UserRepository: userRepository,
	}
}

func (r *FakeSessionRepository) Create(ctx context.Context, input CreateSessionInput) error {
	if r.ReturnError {
		return fmt.Errorf("could not create session %v", input)
	}
	r.lock.Lock()
	defer r.lock.Unlock()
	r.UserIdByToken[input.Token] = input.UserID
	return nil
}

func (r *FakeSessionRepository) GetUserByToken(ctx context.Context, token SessionToken) (u User, err error) {
	userId, ok := r.UserIdByToken[token]
	if !ok {
		return u, &UserDoesNotExistError{}
	}
	return r.UserRepository.GetByID(ctx, userId)
}
