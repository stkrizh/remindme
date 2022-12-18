package channel

import (
	"context"
	"errors"
)

type FakeRepository struct {
	CreateReturnError bool
	Created           []Channel
}

func NewFakeRepository() *FakeRepository {
	return &FakeRepository{}
}

func (r *FakeRepository) Create(ctx context.Context, input CreateInput) (channel Channel, err error) {
	if r.CreateReturnError {
		return channel, errors.New("coulf not create channel")
	}
	channel = Channel{
		ID:         ID(1),
		Settings:   input.Settings,
		CreatedBy:  input.CreatedBy,
		CreatedAt:  input.CreatedAt,
		IsVerified: input.IsVerified,
	}
	r.Created = append(r.Created, channel)
	return channel, nil
}
