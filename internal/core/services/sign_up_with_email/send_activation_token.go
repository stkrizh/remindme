package signupwithemail

import (
	"context"
	"errors"
	e "remindme/internal/core/domain/errors"
	"remindme/internal/core/domain/logging"
	"remindme/internal/core/domain/user"
	"remindme/internal/core/services"
)

type serviceWithActivationTokenSending struct {
	log    logging.Logger
	sender user.ActivationTokenSender
	innner services.Service[Input, Result]
}

func NewWithActivationTokenSending(
	log logging.Logger,
	sender user.ActivationTokenSender,
	innner services.Service[Input, Result],
) services.Service[Input, Result] {
	if log == nil {
		panic(e.NewNilArgumentError("log"))
	}
	if sender == nil {
		panic(e.NewNilArgumentError("sender"))
	}
	if innner == nil {
		panic(e.NewNilArgumentError("innner"))
	}
	return &serviceWithActivationTokenSending{
		log:    log,
		sender: sender,
		innner: innner,
	}
}

func (s *serviceWithActivationTokenSending) Run(ctx context.Context, input Input) (result Result, err error) {
	result, err = s.innner.Run(ctx, input)
	if errors.Is(err, context.Canceled) {
		return result, err
	}
	if err != nil {
		s.log.Info(ctx, "Skip sending activation token.", logging.Entry("err", err))
		return result, err
	}

	err = s.sender.SendToken(ctx, result.User)
	if errors.Is(err, context.Canceled) {
		return result, err
	}
	if err != nil {
		s.log.Error(
			ctx,
			"Could not send activation token.",
			logging.Entry("user", result.User),
			logging.Entry("err", err),
		)
		return result, err
	}

	s.log.Info(
		ctx,
		"Activation token has been sent to the user.",
		logging.Entry("userId", result.User.ID),
		logging.Entry("activationToken", result.User.ActivationToken),
	)
	return result, err
}
