package randomstringgenerator

import (
	"math/rand"
	"remindme/internal/core/domain/channel"
	"remindme/internal/core/domain/user"
	"time"
)

type Generator struct {
	chars []rune
}

func NewGenerator() *Generator {
	rand.Seed(time.Now().UnixNano())
	return &Generator{
		chars: []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"),
	}
}

func (g *Generator) GenerateActivationToken() user.ActivationToken {
	b := make([]rune, 8)
	for i := range b {
		b[i] = g.chars[rand.Intn(len(g.chars))]
	}
	return user.ActivationToken(b)
}

func (g *Generator) GenerateSessionToken() user.SessionToken {
	b := make([]rune, 32)
	for i := range b {
		b[i] = g.chars[rand.Intn(len(g.chars))]
	}
	return user.SessionToken(b)
}

func (g *Generator) GenerateIdentity() user.Identity {
	b := make([]rune, 16)
	for i := range b {
		b[i] = g.chars[rand.Intn(len(g.chars))]
	}
	return user.Identity(b)
}

func (g *Generator) GenerateVerificationToken() channel.VerificationToken {
	b := make([]rune, 8)
	for i := range b {
		b[i] = g.chars[rand.Intn(len(g.chars))]
	}
	return channel.VerificationToken(b)
}
