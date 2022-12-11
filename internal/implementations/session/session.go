package session

import (
	"remindme/internal/core/domain/user"

	"github.com/google/uuid"
)

type UUID struct{}

func NewUUID() *UUID {
	return &UUID{}
}

func (g *UUID) GenerateToken() user.SessionToken {
	return user.SessionToken(uuid.New().String())
}
