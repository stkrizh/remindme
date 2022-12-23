package user

import "context"

type ActivationTokenGenerator interface {
	GenerateActivationToken() ActivationToken
}

type ActivationTokenSender interface {
	SendActivationToken(ctx context.Context, user User) error
}
