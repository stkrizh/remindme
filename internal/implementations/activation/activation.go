package activation

import (
	"math/rand"
	"remindme/internal/domain/user"
	"time"
)

type TokenGenerator struct {
	chars []rune
}

func NewTokenGenerator() *TokenGenerator {
	rand.Seed(time.Now().UnixNano())
	return &TokenGenerator{
		chars: []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"),
	}
}

func (g *TokenGenerator) GenerateToken() user.ActivationToken {
	b := make([]rune, 8)
	for i := range b {
		b[i] = g.chars[rand.Intn(len(g.chars))]
	}
	return user.ActivationToken(b)
}
