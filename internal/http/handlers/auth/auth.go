package auth

import (
	"context"
	"net/http"
	"remindme/internal/core/domain/user"
	"remindme/internal/core/services/auth"
	"strings"
)

const (
	AUTH_TOKEN_PREFIX  = "Bearer "
	AUTH_TOKEN_MAX_LEN = 1024
)

func ParseToken(r *http.Request) (token user.SessionToken, ok bool) {
	header := r.Header.Get("authorization")
	if header == "" {
		return token, false
	}
	parts := strings.SplitN(header, AUTH_TOKEN_PREFIX, 2)
	if len(parts) != 2 {
		return token, false
	}
	if len(parts[1]) > AUTH_TOKEN_MAX_LEN {
		return token, false
	}
	return user.SessionToken(parts[1]), true
}

func SetTokenToContext(r *http.Request) context.Context {
	token, ok := ParseToken(r)
	if !ok {
		return r.Context()
	}
	return context.WithValue(r.Context(), auth.CONTEXT_AUTH_TOKEN_KEY, token)
}
