package user

import "context"

type ActivationTokenGenerator interface {
	GenerateToken() ActivationToken
}

type ActivationTokenSender interface {
	SendToken(ctx context.Context, user User) error
}
