package services

import (
	"remindme/internal/app/deps"
	"remindme/internal/core/domain/channel"
	drl "remindme/internal/core/domain/rate_limiter"
	"remindme/internal/core/services"
	activateuser "remindme/internal/core/services/activate_user"
	"remindme/internal/core/services/auth"
	changepassword "remindme/internal/core/services/change_password"
	createemailchannel "remindme/internal/core/services/create_email_channel"
	createreminder "remindme/internal/core/services/create_reminder"
	createreminderbynlq "remindme/internal/core/services/create_reminder_by_nlq"
	createtelegramchannel "remindme/internal/core/services/create_telegram_channel"
	deletereminder "remindme/internal/core/services/delete_reminder"
	getuserbysessiontoken "remindme/internal/core/services/get_user_by_session_token"
	listuserchannels "remindme/internal/core/services/list_user_channels"
	listuserreminders "remindme/internal/core/services/list_user_reminders"
	loginwithemail "remindme/internal/core/services/log_in_with_email"
	logout "remindme/internal/core/services/log_out"
	ratelimiting "remindme/internal/core/services/rate_limiting"
	resetpassword "remindme/internal/core/services/reset_password"
	schedulereminders "remindme/internal/core/services/schedule_reminders"
	sendpasswordresettoken "remindme/internal/core/services/send_password_reset_token"
	sendreminder "remindme/internal/core/services/send_reminder"
	signupanonymously "remindme/internal/core/services/sign_up_anonymously"
	signupwithemail "remindme/internal/core/services/sign_up_with_email"
	updatereminder "remindme/internal/core/services/update_reminder"
	updatereminderchannels "remindme/internal/core/services/update_reminder_channels"
	updateuser "remindme/internal/core/services/update_user"
	verifyemailchannel "remindme/internal/core/services/verify_email_channel"
	verifytelegramchannel "remindme/internal/core/services/verify_telegram_channel"
)

type Services struct {
	SignUpWithEmail        services.Service[signupwithemail.Input, signupwithemail.Result]
	SignUpAnonymously      services.Service[signupanonymously.Input, signupanonymously.Result]
	ActivateUser           services.Service[activateuser.Input, activateuser.Result]
	LogInWithEmail         services.Service[loginwithemail.Input, loginwithemail.Result]
	LogOut                 services.Service[logout.Input, logout.Result]
	SendPasswordResetToken services.Service[sendpasswordresettoken.Input, sendpasswordresettoken.Result]
	ResetPassword          services.Service[resetpassword.Input, resetpassword.Result]
	ChangePassword         services.Service[changepassword.Input, changepassword.Result]
	GetUserBySessionToken  services.Service[getuserbysessiontoken.Input, getuserbysessiontoken.Result]
	UpdateUser             services.Service[updateuser.Input, updateuser.Result]

	CreateEmailChannel    services.Service[createemailchannel.Input, createemailchannel.Result]
	CreateTelegramChannel services.Service[createtelegramchannel.Input, createtelegramchannel.Result]
	ListUserChannels      services.Service[listuserchannels.Input, listuserchannels.Result]
	VerifyEmailChannel    services.Service[verifyemailchannel.Input, verifyemailchannel.Result]
	VerifyTelegramChannel services.Service[verifytelegramchannel.Input, verifytelegramchannel.Result]

	CreateReminder         services.Service[createreminder.Input, createreminder.Result]
	CreateReminderByNLQ    services.Service[createreminderbynlq.Input, createreminder.Result]
	DeleteReminder         services.Service[deletereminder.Input, deletereminder.Result]
	ListUserReminders      services.Service[listuserreminders.Input, listuserreminders.Result]
	ScheduleReminders      services.Service[schedulereminders.Input, schedulereminders.Result]
	UpdateReminder         services.Service[updatereminder.Input, updatereminder.Result]
	UpdateReminderChannels services.Service[updatereminderchannels.Input, updatereminderchannels.Result]
	SendReminder           services.Service[sendreminder.Input, sendreminder.Result]
}

func InitServices(deps *deps.Deps) *Services {
	s := &Services{}

	s.SignUpWithEmail = signupwithemail.NewWithActivationTokenSending(
		deps.Logger,
		deps.UserActivationTokenSender,
		signupwithemail.New(
			deps.Logger,
			deps.UnitOfWork,
			deps.PasswordHasher,
			deps.UserActivationTokenGenerator,
			deps.Now,
		),
	)
	s.SignUpAnonymously = signupanonymously.New(
		deps.Logger,
		deps.UnitOfWork,
		deps.UserIdentityGenerator,
		deps.UserSessionTokenGenerator,
		deps.Now,
		deps.DefaultAnonymousUserLimits,
	)
	s.ActivateUser = activateuser.New(
		deps.Logger,
		deps.UnitOfWork,
		deps.InternalChannelTokenGenerator,
		deps.Now,
		deps.DefaultUserLimits,
	)
	s.LogInWithEmail = ratelimiting.WithRateLimiting(
		deps.Logger,
		deps.RateLimiter,
		drl.Limit{Interval: drl.Hour, Value: 10},
		loginwithemail.New(
			deps.Logger,
			deps.UserRepository,
			deps.SessionRepository,
			deps.PasswordHasher,
			deps.UserSessionTokenGenerator,
			deps.Now,
		),
	)
	s.LogOut = logout.New(
		deps.Logger,
		deps.SessionRepository,
	)
	s.SendPasswordResetToken = ratelimiting.WithRateLimiting(
		deps.Logger,
		deps.RateLimiter,
		drl.Limit{Interval: drl.Hour, Value: 3},
		sendpasswordresettoken.New(
			deps.Logger,
			deps.UserRepository,
			deps.PasswordResetter,
			deps.PasswordResetTokenSender,
		),
	)
	s.ResetPassword = resetpassword.New(
		deps.Logger,
		deps.UserRepository,
		deps.PasswordResetter,
		deps.PasswordHasher,
	)
	s.ChangePassword = auth.WithAuthentication(
		deps.SessionRepository,
		changepassword.New(
			deps.Logger,
			deps.UserRepository,
			deps.PasswordHasher,
		),
	)
	s.UpdateUser = auth.WithAuthentication(
		deps.SessionRepository,
		updateuser.New(
			deps.Logger,
			deps.UserRepository,
		),
	)
	s.GetUserBySessionToken = auth.WithAuthentication(
		deps.SessionRepository,
		getuserbysessiontoken.New(),
	)

	s.CreateEmailChannel = auth.WithAuthentication(
		deps.SessionRepository,
		createemailchannel.NewWithVerificationTokenSending(
			deps.Logger,
			channel.NewFakeVerificationTokenSender(),
			createemailchannel.New(
				deps.Logger,
				deps.UnitOfWork,
				deps.ChannelVerificationTokenGenerator,
				deps.Now,
			),
		),
	)
	s.CreateTelegramChannel = auth.WithAuthentication(
		deps.SessionRepository,
		createtelegramchannel.New(
			deps.Logger,
			deps.UnitOfWork,
			deps.ChannelVerificationTokenGenerator,
			deps.Now,
		),
	)
	s.ListUserChannels = auth.WithAuthentication(
		deps.SessionRepository,
		listuserchannels.New(
			deps.Logger,
			deps.ChannelRepository,
		),
	)
	s.VerifyEmailChannel = auth.WithAuthentication(
		deps.SessionRepository,
		ratelimiting.WithRateLimiting(
			deps.Logger,
			deps.RateLimiter,
			drl.Limit{Interval: drl.Minute, Value: 5},
			verifyemailchannel.New(
				deps.Logger,
				deps.ChannelRepository,
				deps.Now,
			),
		),
	)
	s.VerifyTelegramChannel = verifytelegramchannel.New(
		deps.Logger,
		deps.ChannelRepository,
		deps.Now,
	)

	s.CreateReminder = auth.WithAuthentication(
		deps.SessionRepository,
		createreminder.New(
			deps.Logger,
			deps.UnitOfWork,
			deps.ReminderScheduler,
			deps.Now,
		),
	)
	s.CreateReminderByNLQ = auth.WithAuthentication(
		deps.SessionRepository,
		createreminderbynlq.New(
			deps.Logger,
			deps.ReminderNLQParser,
			deps.ChannelRepository,
			deps.Now,
			createreminder.New(
				deps.Logger,
				deps.UnitOfWork,
				deps.ReminderScheduler,
				deps.Now,
			),
		),
	)
	s.DeleteReminder = auth.WithAuthentication(
		deps.SessionRepository,
		deletereminder.New(
			deps.Logger,
			deps.UnitOfWork,
			deps.Now,
		),
	)
	s.ListUserReminders = auth.WithAuthentication(
		deps.SessionRepository,
		listuserreminders.New(
			deps.Logger,
			deps.ReminderRepository,
		),
	)
	s.ScheduleReminders = schedulereminders.New(
		deps.Logger,
		deps.UnitOfWork,
		deps.ReminderScheduler,
		deps.Now,
	)
	s.UpdateReminder = auth.WithAuthentication(
		deps.SessionRepository,
		updatereminder.New(
			deps.Logger,
			deps.UnitOfWork,
			deps.ReminderScheduler,
			deps.Now,
		),
	)
	s.UpdateReminderChannels = auth.WithAuthentication(
		deps.SessionRepository,
		updatereminderchannels.New(
			deps.Logger,
			deps.UnitOfWork,
			deps.Now,
		),
	)
	s.SendReminder = sendreminder.NewSendService(
		deps.Logger,
		deps.ReminderRepository,
		deps.ReminderSender,
		deps.Now,
		sendreminder.NewPrepareService(
			deps.Logger,
			deps.UnitOfWork,
			deps.Now,
		),
	)

	return s
}
