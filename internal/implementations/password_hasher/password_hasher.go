package passwordhasher

import (
	"remindme/internal/domain/user"

	"golang.org/x/crypto/bcrypt"
)

type Bcrypt struct {
	secret string
	cost   int
}

func NewBcrypt(secret string, cost int) *Bcrypt {
	return &Bcrypt{secret: secret, cost: cost}
}

func (h *Bcrypt) HashPassword(password user.RawPassword) (hash user.PasswordHash, err error) {
	bcryptHash, err := bcrypt.GenerateFromPassword([]byte(string(password)+h.secret), h.cost)
	if err != nil {
		return hash, err
	}
	return user.PasswordHash(bcryptHash), nil
}

func (h *Bcrypt) ValidatePassword(password user.RawPassword, hash user.PasswordHash) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(string(password)+h.secret))
	return err == nil
}
