package captcha

import (
	"context"
	"net/http"
	"remindme/internal/core/services/captcha"
)

func SetCaptchaTokenToContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("X-Captcha-Token")
		if token != "" {
			ctx := context.WithValue(r.Context(), captcha.CONTEXT_CAPTCHA_TOKEN_KEY, captcha.CaptchaToken(token))
			r = r.WithContext(ctx)
		}
		next.ServeHTTP(w, r)
	})
}
