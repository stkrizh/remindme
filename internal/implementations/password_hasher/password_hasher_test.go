package passwordhasher

import (
	"fmt"
	"remindme/internal/domain/user"
	"testing"
)

func TestPasswordValid(t *testing.T) {
	type testcase struct {
		ix       int
		secret   string
		cost     int
		password string
	}
	cases := []testcase{
		{ix: 1, secret: "test", cost: 5, password: "test"},
		{ix: 2, secret: "", cost: 5, password: ""},
		{ix: 3, secret: "a", cost: 7, password: "password password"},
		{ix: 4, secret: "   b   ", cost: 10, password: "   test   "},
	}
	for _, c := range cases {
		t.Run(fmt.Sprint(c.ix), func(t *testing.T) {
			h := NewBcrypt(c.secret, c.cost)
			hash, err := h.HashPassword(user.RawPassword(c.password))
			if hash == user.PasswordHash("") {
				t.Fatal("hash must not be empty")
			}
			if err != nil {
				t.Fatalf("could not hash password: %v, %v", c.password, err)
			}
			if !h.ValidatePassword(user.RawPassword(c.password), hash) {
				t.Fatalf("password check failed: %v", c.password)
			}
		})
	}
}

func TestPasswordInvalid(t *testing.T) {
	type testcase struct {
		ix              int
		secretToHash    string
		secretToCheck   string
		cost            int
		passwordToHash  string
		passwordToCheck string
	}
	cases := []testcase{
		{
			ix:              1,
			secretToHash:    "test",
			secretToCheck:   "test",
			cost:            5,
			passwordToHash:  "test",
			passwordToCheck: "test ",
		},
		{
			ix:              2,
			secretToHash:    "test",
			secretToCheck:   "test ",
			cost:            5,
			passwordToHash:  "test",
			passwordToCheck: "test",
		},
		{
			ix:              3,
			secretToHash:    "",
			secretToCheck:   "",
			cost:            5,
			passwordToHash:  "",
			passwordToCheck: " ",
		},
		{
			ix:              4,
			secretToHash:    "",
			secretToCheck:   " ",
			cost:            8,
			passwordToHash:  "",
			passwordToCheck: "",
		},
		{
			ix:              5,
			secretToHash:    "a",
			secretToCheck:   "a",
			cost:            10,
			passwordToHash:  "password password",
			passwordToCheck: " password password",
		},
		{
			ix:              6,
			secretToHash:    "   b   ",
			secretToCheck:   "   b   ",
			cost:            8,
			passwordToHash:  "   test   ",
			passwordToCheck: "   tost   ",
		},
	}
	for _, c := range cases {
		t.Run(fmt.Sprint(c.ix), func(t *testing.T) {
			h := NewBcrypt(c.secretToHash, c.cost)
			hash, err := h.HashPassword(user.RawPassword(c.passwordToHash))
			if hash == user.PasswordHash("") {
				t.Fatal("hash must not be empty")
			}
			if err != nil {
				t.Fatalf("could not hash password: %v, %v", c.passwordToHash, err)
			}

			h = NewBcrypt(c.secretToCheck, c.cost)
			if h.ValidatePassword(user.RawPassword(c.passwordToCheck), hash) {
				t.Fatalf("password check passed: %v, %v", c.passwordToHash, c.passwordToCheck)
			}
		})
	}
}
