package user

type IdentityGenerator interface {
	GenerateIdentity() Identity
}
