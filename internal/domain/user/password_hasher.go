package user

type PasswordHasher interface {
	HashPassword(password RawPassword) PasswordHash
	ValidatePassword(password RawPassword, hash PasswordHash) bool
}
