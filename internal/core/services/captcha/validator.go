package captcha

import "context"

type CaptchaToken string

func (t CaptchaToken) IsZero() bool {
	return string(t) == ""
}

type CaptchaValidator interface {
	ValidateCaptchaToken(ctx context.Context, token CaptchaToken) bool
}
