package identity

import (
	"remindme/internal/core/domain/user"
	"testing"
)

func TestIdentityGenerator(t *testing.T) {
	generator := NewUUID()
	identities := make(map[user.Identity]struct{})
	for i := 0; i < 100; i++ {
		identity := generator.GenerateIdentity()
		if string(identity) == "" {
			t.Fatal("identity must not be empty")
		}
		if _, ok := identities[identity]; ok {
			t.Fatalf("identity %v already exists (%v)", identity, identities)
		}
		identities[identity] = struct{}{}
	}
}
