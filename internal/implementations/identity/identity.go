package identity

import (
	"remindme/internal/core/domain/user"

	"github.com/google/uuid"
)

type UUID struct{}

func NewUUID() *UUID {
	return &UUID{}
}

func (g *UUID) GenerateIdentity() user.Identity {
	return user.Identity(uuid.New().String())
}
