package user

import (
	"context"
	"crypto/md5"
	"fmt"
	"io"
	c "remindme/internal/core/domain/common"
	"sync"
	"time"
)

type FakeActivationTokenSender struct {
	Sent        []User
	ReturnError bool
	lock        sync.Mutex
}

func NewFakeActivationTokenSender() *FakeActivationTokenSender {
	return &FakeActivationTokenSender{}
}

func (s *FakeActivationTokenSender) SendActivationToken(ctx context.Context, user User) error {
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

func (g *FakeActivationTokenGenerator) GenerateActivationToken() ActivationToken {
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

type FakeIdentityGenerator struct {
	Identity string
}

func NewFakeIdentityGenerator(identity string) *FakeIdentityGenerator {
	return &FakeIdentityGenerator{Identity: identity}
}

type FakeSessionTokenGenerator struct {
	Token string
}

func NewFakeSessionTokenGenerator(token string) *FakeSessionTokenGenerator {
	return &FakeSessionTokenGenerator{Token: token}
}

func (g *FakeSessionTokenGenerator) GenerateSessionToken() SessionToken {
	return SessionToken(g.Token)
}

func (g *FakeIdentityGenerator) GenerateIdentity() Identity {
	return Identity(g.Identity)
}

type FakeUserRepository struct {
	Users       []User
	ReturnError bool
	lock        sync.Mutex
}

func NewFakeUserRepository() *FakeUserRepository {
	return &FakeUserRepository{Users: make([]User, 0, 10)}
}

func (r *FakeUserRepository) Create(ctx context.Context, input CreateUserInput) (u User, err error) {
	if r.ReturnError {
		return u, fmt.Errorf("could not create user %v", input)
	}
	r.lock.Lock()
	defer r.lock.Unlock()
	maxID := ID(0)
	for _, u := range r.Users {
		if input.Email.IsPresent && u.Email == input.Email {
			return u, ErrEmailAlreadyExists
		}
		maxID = u.ID
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
	r.Users = append(r.Users, u)
	return u, nil
}

func (r *FakeUserRepository) GetByID(ctx context.Context, id ID) (u User, err error) {
	r.lock.Lock()
	defer r.lock.Unlock()
	for _, u := range r.Users {
		if u.ID == id {
			return u, nil
		}
	}
	return u, ErrUserDoesNotExist
}

func (r *FakeUserRepository) GetByEmail(ctx context.Context, email c.Email) (u User, err error) {
	r.lock.Lock()
	defer r.lock.Unlock()
	for _, u := range r.Users {
		if u.Email.IsPresent && u.Email.Value == email {
			return u, nil
		}
	}
	return u, ErrUserDoesNotExist
}

func (r *FakeUserRepository) Activate(ctx context.Context, token ActivationToken, at time.Time) (u User, err error) {
	r.lock.Lock()
	defer r.lock.Unlock()
	for ix, u := range r.Users {
		if !u.IsActive() && u.ActivationToken.IsPresent && u.ActivationToken.Value == token {
			r.Users[ix].ActivatedAt = c.NewOptional(at, true)
			r.Users[ix].ActivationToken = c.NewOptional(ActivationToken(""), false)
			return r.Users[ix], nil
		}
	}
	return u, ErrInvalidActivationToken
}

func (r *FakeUserRepository) SetPassword(ctx context.Context, id ID, password PasswordHash) error {
	r.lock.Lock()
	defer r.lock.Unlock()
	for ix, u := range r.Users {
		if u.ID == id {
			r.Users[ix].PasswordHash = c.NewOptional(password, true)
			return nil
		}
	}
	return ErrUserDoesNotExist
}

func (r *FakeUserRepository) Update(ctx context.Context, input UpdateUserInput) (u User, err error) {
	r.lock.Lock()
	defer r.lock.Unlock()
	for ix, u := range r.Users {
		if u.ID == input.ID {
			if input.DoTimeZoneUpdate {
				r.Users[ix].TimeZone = input.TimeZone
			}
			return r.Users[ix], nil
		}
	}
	return u, ErrUserDoesNotExist
}

type FakeSessionRepository struct {
	UserIdByToken  map[SessionToken]ID
	UserRepository UserRepository
	ReturnError    bool
	lock           sync.Mutex
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
	r.lock.Lock()
	defer r.lock.Unlock()
	userId, ok := r.UserIdByToken[token]
	if !ok {
		return u, ErrUserDoesNotExist
	}
	return r.UserRepository.GetByID(ctx, userId)
}

func (r *FakeSessionRepository) Delete(ctx context.Context, token SessionToken) (ID, error) {
	r.lock.Lock()
	defer r.lock.Unlock()
	userID, ok := r.UserIdByToken[token]
	if !ok {
		return ID(0), ErrSessionDoesNotExist
	}
	delete(r.UserIdByToken, token)
	return userID, nil
}

type FakeLimitsRepository struct {
	ReturnError bool
	Created     []Limits
	Limits      Limits
	lock        sync.Mutex
}

func NewFakeLimitsRepository() *FakeLimitsRepository {
	return &FakeLimitsRepository{}
}

func (r *FakeLimitsRepository) Create(ctx context.Context, input CreateLimitsInput) (l Limits, err error) {
	if r.ReturnError {
		return l, fmt.Errorf("could not create limits %v", input)
	}
	r.lock.Lock()
	defer r.lock.Unlock()
	l.EmailChannelCount = input.Limits.EmailChannelCount
	l.TelegramChannelCount = input.Limits.TelegramChannelCount
	r.Created = append(r.Created, l)
	return l, nil
}

func (r *FakeLimitsRepository) GetUserLimits(ctx context.Context, userID ID) (l Limits, err error) {
	if r.ReturnError {
		return l, fmt.Errorf("could not get user limits")
	}
	return r.Limits, nil
}

func (r *FakeLimitsRepository) GetUserLimitsWithLock(ctx context.Context, userID ID) (l Limits, err error) {
	if r.ReturnError {
		return l, fmt.Errorf("could not get user limits")
	}
	return r.Limits, nil
}

type FakePasswordResetter struct {
	Token         PasswordResetToken
	UserID        ID
	IsUserIDValid bool
	IsValid       bool
}

func NewFakePasswordResetter(token string, userID ID, isUserIDValid bool, isValid bool) *FakePasswordResetter {
	return &FakePasswordResetter{
		Token:         PasswordResetToken(token),
		UserID:        userID,
		IsUserIDValid: isUserIDValid,
		IsValid:       isValid,
	}
}

func (r *FakePasswordResetter) GenerateToken(user User) PasswordResetToken {
	return r.Token
}

func (r *FakePasswordResetter) GetUserID(token PasswordResetToken) (ID, bool) {
	return r.UserID, r.IsUserIDValid
}

func (r *FakePasswordResetter) ValidateToken(user User, token PasswordResetToken) bool {
	return r.IsValid
}

type FakePasswordResetTokenSender struct {
	Sent        []PasswordResetToken
	SentTo      []User
	ReturnError bool
	lock        sync.Mutex
}

func NewFakePasswordResetTokenSender() *FakePasswordResetTokenSender {
	return &FakePasswordResetTokenSender{}
}

func (s *FakePasswordResetTokenSender) SendPasswordResetToken(
	ctx context.Context,
	user User,
	token PasswordResetToken,
) error {
	if s.ReturnError {
		return fmt.Errorf("could not send password reset token")
	}
	s.lock.Lock()
	defer s.lock.Unlock()
	s.Sent = append(s.Sent, token)
	s.SentTo = append(s.SentTo, user)
	return nil
}
