package channel

import (
	"context"
)

type VerificationTokenSender interface {
	SendVerificationToken(ctx context.Context, token VerificationToken, channel Channel) error
}

type VerificationTokenGenerator interface {
	GenerateVerificationToken() VerificationToken
}
