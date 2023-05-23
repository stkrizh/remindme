package passwordresetter

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"io"
	"math/rand"
	"remindme/internal/core/domain/user"
	"strconv"
	"strings"
	"time"
)

var saltChars = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

type HMAC struct {
	secretKey     []byte
	validDuration time.Duration
	now           func() time.Time
}

func NewHMAC(secretKey string, validDuration time.Duration, now func() time.Time) *HMAC {
	return &HMAC{
		secretKey:     []byte(secretKey),
		validDuration: validDuration,
		now:           now,
	}
}

func (h *HMAC) GenerateToken(u user.User) user.PasswordResetToken {
	nowTs := h.now().Unix()
	salt := h.getRandomSalt()
	mac := h.getMac(u.ID, u.PasswordHash.Value, nowTs, salt)
	b64 := base64.RawURLEncoding.EncodeToString([]byte(fmt.Sprintf("%d-%d-%s-%s", u.ID, nowTs, salt, mac)))
	return user.PasswordResetToken(b64)
}

func (h *HMAC) ValidateToken(u user.User, token user.PasswordResetToken) bool {
	decodedToken, err := base64.RawURLEncoding.DecodeString(string(token))
	if err != nil {
		return false
	}
	parts := strings.SplitN(string(decodedToken), "-", 4)
	if len(parts) != 4 {
		return false
	}
	ts, err := strconv.Atoi(parts[1])
	if err != nil {
		return false
	}
	actualDuration := time.Duration((h.now().Unix() - int64(ts))) * time.Second
	if actualDuration > h.validDuration {
		return false
	}
	salt := parts[2]
	mac := parts[3]
	expectedMac := h.getMac(u.ID, u.PasswordHash.Value, int64(ts), salt)
	return subtle.ConstantTimeCompare([]byte(mac), []byte(expectedMac)) == 1
}

func (h *HMAC) GetUserID(token user.PasswordResetToken) (userID user.ID, ok bool) {
	decodedToken, err := base64.RawURLEncoding.DecodeString(string(token))
	if err != nil {
		return userID, false
	}
	parts := strings.SplitN(string(decodedToken), "-", 4)
	if len(parts) != 4 {
		return userID, false
	}
	rawUserID, err := strconv.Atoi(parts[0])
	if err != nil {
		return userID, false
	}
	return user.ID(rawUserID), true
}

func (h *HMAC) getMac(userID user.ID, passwordHash user.PasswordHash, ts int64, salt string) string {
	hasher := hmac.New(sha256.New, h.secretKey)
	io.WriteString(hasher, fmt.Sprintf("%d-%d-%s-%s", userID, ts, salt, string(passwordHash)))
	return fmt.Sprintf("%x", hasher.Sum(nil))
}

func (h *HMAC) getRandomSalt() string {
	b := make([]rune, 5)
	for i := range b {
		b[i] = saltChars[rand.Intn(len(saltChars))]
	}
	return string(b)
}
