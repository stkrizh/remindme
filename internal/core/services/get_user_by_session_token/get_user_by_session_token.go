package getuserbysessiontoken

import (
	"context"
	"remindme/internal/core/domain/user"
	"remindme/internal/core/services"
	"remindme/internal/core/services/auth"
)

type Input struct {
	User user.User
}

func (i Input) WithAuthenticatedUser(u user.User) auth.Input {
	i.User = u
	return i
}

type Result struct {
	User user.User
}

type service struct{}

func New() services.Service[Input, Result] {
	return &service{}
}

func (s *service) Run(ctx context.Context, input Input) (result Result, err error) {
	return Result(input), nil
}
