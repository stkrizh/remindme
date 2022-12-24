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

func SetAuthTokenToContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, ok := ParseToken(r)
		if ok {
			ctx := context.WithValue(r.Context(), auth.CONTEXT_AUTH_TOKEN_KEY, token)
			r = r.WithContext(ctx)
		}
		next.ServeHTTP(w, r)
	})
}
