package user

type PasswordHasher interface {
	HashPassword(password RawPassword) (PasswordHash, error)
	ValidatePassword(password RawPassword, hash PasswordHash) bool
}
