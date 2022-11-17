package service

import "remindme/internal/domain/user"

func PrintUser() *user.User {
	u := user.User{Id: user.UserId(42)}
	return &u
}
