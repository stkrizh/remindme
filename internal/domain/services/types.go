package services

import "remindme/internal/domain/user"

type SignUpWithEmailInput struct {
	Email    user.Email
	Password user.RawPassword
}

type SignUpWithEmailResult struct {
	User user.User
}
