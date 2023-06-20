package user

import "context"

type PasswordResetToken string

type PasswordResetter interface {
	GenerateToken(user User) PasswordResetToken
	GetUserID(token PasswordResetToken) (ID, bool)
	ValidateToken(user User, token PasswordResetToken) bool
}

type PasswordResetTokenSender interface {
	SendPasswordResetToken(ctx context.Context, user User, token PasswordResetToken) error
}
