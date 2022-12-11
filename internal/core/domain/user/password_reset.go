package user

import "context"

type PasswordResetToken string

type PasswordResetter interface {
	GenerateToken(user User) PasswordResetToken
	ValidateToken(user User, token PasswordResetToken) bool
}

type PasswordResetTokenSender interface {
	SendToken(ctx context.Context, user User, token PasswordResetToken) error
}
