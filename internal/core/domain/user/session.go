package user

type SessionTokenGenerator interface {
	GenerateSessionToken() SessionToken
}
