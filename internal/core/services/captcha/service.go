package captcha

import (
	"context"
	e "remindme/internal/core/domain/errors"
	"remindme/internal/core/services"
)

type contextCaptchaToken string

const CONTEXT_CAPTCHA_TOKEN_KEY = contextCaptchaToken("captchaToken")

type service[T any, S any] struct {
	validator CaptchaValidator
	inner     services.Service[T, S]
}

func WithCaptcha[T any, S any](
	validator CaptchaValidator,
	inner services.Service[T, S],
) services.Service[T, S] {
	if validator == nil {
		panic(e.NewNilArgumentError("validator"))
	}
	if inner == nil {
		panic(e.NewNilArgumentError("inner"))
	}
	return &service[T, S]{
		validator: validator,
		inner:     inner,
	}
}

func (s *service[T, S]) Run(ctx context.Context, input T) (result S, err error) {
	token, _ := ctx.Value(CONTEXT_CAPTCHA_TOKEN_KEY).(CaptchaToken)
	if !s.validator.ValidateCaptchaToken(ctx, token) {
		return result, ErrInvalidCaptcha
	}
	return s.inner.Run(ctx, input)
}
