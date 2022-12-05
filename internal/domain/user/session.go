package user

type SessionTokenGenerator interface {
	GenerateToken() SessionToken
}
