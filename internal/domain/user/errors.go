package user

import "fmt"

type EmailAlreadyExistsError struct {
	Email Email
}

func (e *EmailAlreadyExistsError) Error() string {
	return fmt.Sprintf("user with email %s already exists", e.Email)
}

type IdentityAlreadyExistsError struct {
	Identity Identity
}

func (e *IdentityAlreadyExistsError) Error() string {
	return fmt.Sprintf("user with identity %s already exists", e.Identity)
}

type ActivationTokenAlreadyExistsError struct {
	ActivationToken ActivationToken
}

func (e *ActivationTokenAlreadyExistsError) Error() string {
	return fmt.Sprintf("user with activation token %s already exists", e.ActivationToken)
}

type UserDoesNotExistError struct{}

func (e *UserDoesNotExistError) Error() string {
	return "user does not exists"
}
