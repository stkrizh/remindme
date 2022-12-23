package createemailchannel

import (
	"context"
	"errors"
	"remindme/internal/core/domain/channel"
	e "remindme/internal/core/domain/errors"
	"remindme/internal/core/domain/logging"
	"remindme/internal/core/services"
)

type sendVerificationTokenService struct {
	log    logging.Logger
	sender channel.VerificationTokenSender
	inner  services.Service[Input, Result]
}

func NewWithVerificationTokenSending(
	log logging.Logger,
	sender channel.VerificationTokenSender,
	inner services.Service[Input, Result],
) services.Service[Input, Result] {
	if log == nil {
		panic(e.NewNilArgumentError("log"))
	}
	if sender == nil {
		panic(e.NewNilArgumentError("sender"))
	}
	if inner == nil {
		panic(e.NewNilArgumentError("inner"))
	}
	return &sendVerificationTokenService{
		log:    log,
		sender: sender,
		inner:  inner,
	}
}

func (s *sendVerificationTokenService) Run(ctx context.Context, input Input) (result Result, err error) {
	result, err = s.inner.Run(ctx, input)
	if errors.Is(err, context.Canceled) {
		return result, err
	}
	if err != nil {
		s.log.Warning(
			ctx,
			"Inner service returned an error, skip token sending.",
			logging.Entry("email", input.Email),
			logging.Entry("userID", input.UserID),
			logging.Entry("err", err),
		)
		return result, err
	}

	err = s.sender.SendVerificationToken(ctx, result.VerificationToken, result.Channel)
	if errors.Is(err, context.Canceled) {
		return result, err
	}
	if err != nil {
		s.log.Error(
			ctx,
			"Could not send email channel verification token.",
			logging.Entry("email", input.Email),
			logging.Entry("userID", input.UserID),
			logging.Entry("err", err),
		)
		return result, err
	}

	s.log.Info(
		ctx,
		"Email channel verification token has been sent.",
		logging.Entry("email", input.Email),
		logging.Entry("userID", input.UserID),
		logging.Entry("channelID", result.Channel.ID),
	)
	return result, nil
}
