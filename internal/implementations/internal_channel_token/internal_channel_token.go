package internalchanneltoken

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"math/rand"
	"remindme/internal/core/domain/channel"
	"strings"
)

var saltChars = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

type HMAC struct {
	secretKey []byte
}

func NewHMAC(secretKey string) *HMAC {
	return &HMAC{
		secretKey: []byte(secretKey),
	}
}

func (h *HMAC) GenerateInternalChannelToken() channel.InternalChannelToken {
	salt := h.getRandomSalt()
	mac := h.getMac(salt)
	b64 := base64.RawURLEncoding.EncodeToString([]byte(fmt.Sprintf("%s-%s", salt, mac)))
	return channel.InternalChannelToken(b64)
}

func (h *HMAC) ValidateInternalChannelToken(token channel.InternalChannelToken) bool {
	decodedToken, err := base64.RawURLEncoding.DecodeString(string(token))
	if err != nil {
		return false
	}
	parts := strings.SplitN(string(decodedToken), "-", 2)
	if len(parts) != 2 {
		return false
	}
	salt := parts[0]
	mac := parts[1]
	expectedMac := h.getMac(salt)
	return expectedMac == mac
}

func (h *HMAC) getMac(salt string) string {
	hasher := hmac.New(sha256.New, h.secretKey)
	io.WriteString(hasher, salt)
	return fmt.Sprintf("%x", hasher.Sum(nil))
}

func (h *HMAC) getRandomSalt() string {
	b := make([]rune, 8)
	for i := range b {
		b[i] = saltChars[rand.Intn(len(saltChars))]
	}
	return string(b)
}
