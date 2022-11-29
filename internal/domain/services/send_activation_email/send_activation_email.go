package sendactivationemail

import (
	"context"
	"remindme/internal/domain/logging"
	"remindme/internal/domain/services"
	"remindme/internal/domain/user"
)

type service struct {
	log           logging.Logger
	sender        user.ActivationTokenSender
	signUpService services.Service[services.SignUpWithEmailInput, services.SignUpWithEmailResult]
}

func New(
	log logging.Logger,
	sender user.ActivationTokenSender,
	signUpService services.Service[services.SignUpWithEmailInput, services.SignUpWithEmailResult],
) services.Service[services.SignUpWithEmailInput, services.SignUpWithEmailResult] {
	return &service{
		log:           log,
		sender:        sender,
		signUpService: signUpService,
	}
}

func (s *service) Run(
	ctx context.Context,
	input services.SignUpWithEmailInput,
) (result services.SignUpWithEmailResult, err error) {
	result, err = s.signUpService.Run(ctx, input)
	if err != nil {
		s.log.Info(ctx, "Skip sending activation token.", logging.Entry("err", err))
		return result, err
	}
	if err = s.sender.SendToken(ctx, result.User); err != nil {
		s.log.Error(
			ctx,
			"Could not send activation token.",
			logging.Entry("user", result.User),
			logging.Entry("err", err),
		)
		return result, &user.ActivationTokenSendingError{}
	}
	s.log.Info(ctx, "Activation token has been sent to the user.", logging.Entry("userId", result.User.ID))
	return result, err
}
