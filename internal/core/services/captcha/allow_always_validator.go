package captcha

import "context"

type AllowAlwaysCaptchaValidator struct{}

func NewAllowAlwaysCaptchaValidator() *AllowAlwaysCaptchaValidator {
	return &AllowAlwaysCaptchaValidator{}
}

func (v *AllowAlwaysCaptchaValidator) ValidateCaptchaToken(ctx context.Context, token CaptchaToken) bool {
	return true
}
