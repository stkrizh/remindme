package randomstringgenerator

import (
	"remindme/internal/core/domain/user"
	"testing"
)

func TestIdentityGenerator(t *testing.T) {
	generator := NewGenerator()
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

func TestActivationTokenGenerator(t *testing.T) {
	generator := NewGenerator()
	activationTokens := make(map[user.ActivationToken]struct{})
	for i := 0; i < 100; i++ {
		activationToken := generator.GenerateActivationToken()
		if string(activationToken) == "" {
			t.Fatal("activationToken must not be empty")
		}
		if _, ok := activationTokens[activationToken]; ok {
			t.Fatalf("activationToken %v already exists (%v)", activationToken, activationTokens)
		}
		activationTokens[activationToken] = struct{}{}
	}
}

func TestSessionTokenGenerator(t *testing.T) {
	generator := NewGenerator()
	sessionTokens := make(map[user.SessionToken]struct{})
	for i := 0; i < 100; i++ {
		sessionToken := generator.GenerateSessionToken()
		if string(sessionToken) == "" {
			t.Fatal("sessionToken must not be empty")
		}
		if _, ok := sessionTokens[sessionToken]; ok {
			t.Fatalf("sessionToken %v already exists (%v)", sessionToken, sessionTokens)
		}
		sessionTokens[sessionToken] = struct{}{}
	}
}
