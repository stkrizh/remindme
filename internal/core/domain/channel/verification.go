package channel

import (
	"context"
)

type VerificationTokenSender interface {
	SendToken(ctx context.Context, token VerificationToken, channel Channel) error
}

type VerificationTokenGenerator interface {
	GenerateToken() VerificationToken
}
