package session

import (
	"remindme/internal/domain/user"
	"testing"
)

func TestIdentityGenerator(t *testing.T) {
	generator := NewUUID()
	tokens := make(map[user.SessionToken]struct{})
	for i := 0; i < 100; i++ {
		token := generator.GenerateToken()
		if string(token) == "" {
			t.Fatal("token must not be empty")
		}
		if _, ok := tokens[token]; ok {
			t.Fatalf("token %v already exists (%v)", token, tokens)
		}
		tokens[token] = struct{}{}
	}
}
