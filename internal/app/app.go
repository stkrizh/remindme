package app

import (
	"fmt"
	"net/http"
	"remindme/internal/app/deps"
	"remindme/internal/app/services"
	"remindme/internal/core/domain/channel"
	"remindme/internal/http/handlers/auth"
	activateuser "remindme/internal/http/handlers/auth/activate_user"
	loginwithemail "remindme/internal/http/handlers/auth/log_in_with_email"
	logout "remindme/internal/http/handlers/auth/log_out"
	resetpassword "remindme/internal/http/handlers/auth/reset_password"
	sendpasswordresettoken "remindme/internal/http/handlers/auth/send_password_reset_token"
	signupanonymously "remindme/internal/http/handlers/auth/sign_up_anonymously"
	signupwithemail "remindme/internal/http/handlers/auth/sign_up_with_email"
	"remindme/internal/http/handlers/captcha"
	createemailchannel "remindme/internal/http/handlers/channels/create_email_channel"
	createtlgchannel "remindme/internal/http/handlers/channels/create_telegram_channel"
	internalchannelevents "remindme/internal/http/handlers/channels/internal_channel_events"
	listuserchannels "remindme/internal/http/handlers/channels/list_user_channels"
	verifyemailchannel "remindme/internal/http/handlers/channels/verify_email_channel"
	cancelreminder "remindme/internal/http/handlers/reminders/cancel_reminder"
	createreminder "remindme/internal/http/handlers/reminders/create_reminder"
	createreminderbynlq "remindme/internal/http/handlers/reminders/create_reminder_by_nlq"
	listuserreminders "remindme/internal/http/handlers/reminders/list_user_reminders"
	updatereminder "remindme/internal/http/handlers/reminders/update_reminder"
	updatereminderchannels "remindme/internal/http/handlers/reminders/update_reminder_channels"
	telegram "remindme/internal/http/handlers/telegram"
	changepassword "remindme/internal/http/handlers/user/change_password"
	limitforactivereminders "remindme/internal/http/handlers/user/limit_for_active_reminders"
	limitforchannels "remindme/internal/http/handlers/user/limit_for_channels"
	limitforsentreminders "remindme/internal/http/handlers/user/limit_for_sent_reminders"
	me "remindme/internal/http/handlers/user/me"
	updateuser "remindme/internal/http/handlers/user/update_user"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
)

func InitHttpServer(deps *deps.Deps, s *services.Services) *http.Server {
	isTestMode := deps.Config.IsTestMode

	authRouter := chi.NewRouter()
	authRouter.Method(http.MethodPost, "/signup", signupwithemail.New(s.SignUpWithEmail, isTestMode))
	authRouter.Method(http.MethodPost, "/signup/anonymously", signupanonymously.New(s.SignUpAnonymously))
	authRouter.Method(http.MethodPost, "/activate", activateuser.New(s.ActivateUser))
	authRouter.Method(http.MethodPost, "/login", loginwithemail.New(s.LogInWithEmail))
	authRouter.Method(http.MethodPost, "/logout", logout.New(s.LogOut))
	authRouter.Method(
		http.MethodPost,
		"/password_reset/token",
		sendpasswordresettoken.New(s.SendPasswordResetToken, isTestMode),
	)
	authRouter.Method(http.MethodPut, "/password_reset", resetpassword.New(s.ResetPassword))

	profileRouter := chi.NewRouter()
	profileRouter.Use(auth.SetAuthTokenToContext)
	profileRouter.Method(http.MethodGet, "/me", me.New(s.GetUserBySessionToken))
	profileRouter.Method(http.MethodPatch, "/me", updateuser.New(s.UpdateUser))
	profileRouter.Method(http.MethodPut, "/password", changepassword.New(s.ChangePassword))
	profileRouter.Method(
		http.MethodGet,
		"/limit/reminders/active",
		limitforactivereminders.New(s.GetLimitForActiveReminders),
	)
	profileRouter.Method(
		http.MethodGet,
		"/limit/reminders/sent",
		limitforsentreminders.New(s.GetLimitForSentReminders),
	)
	profileRouter.Method(http.MethodGet, "/limit/channels", limitforchannels.New(s.GetLimitForChannels))

	channelsRouter := chi.NewRouter()
	channelsRouter.Use(auth.SetAuthTokenToContext)
	channelsRouter.Method(http.MethodGet, "/", listuserchannels.New(s.ListUserChannels))
	channelsRouter.Method(http.MethodPost, "/email", createemailchannel.New(s.CreateEmailChannel, isTestMode))
	channelsRouter.Method(
		http.MethodPost,
		"/telegram",
		createtlgchannel.New(s.CreateTelegramChannel, channel.TelegramBot(deps.Config.TelegramBots[0])),
	)
	channelsRouter.Method(
		http.MethodPut,
		"/{channelID:[0-9]+}/verification",
		verifyemailchannel.New(s.VerifyEmailChannel),
	)
	channelsRouter.Method(
		http.MethodGet,
		"/internal/events",
		internalchannelevents.New(deps.Logger, deps.SseServer, deps.InternalChannelTokenValidator),
	)
	// TODO: delete channel

	reminderRouter := chi.NewRouter()
	reminderRouter.Use(auth.SetAuthTokenToContext)
	reminderRouter.Method(http.MethodPost, "/", createreminder.New(s.CreateReminder))
	reminderRouter.Method(http.MethodPost, "/nlq", createreminderbynlq.New(s.CreateReminderByNLQ))
	reminderRouter.Method(http.MethodGet, "/", listuserreminders.New(s.ListUserReminders))
	reminderRouter.Method(http.MethodDelete, "/{reminderID:[0-9]+}", cancelreminder.New(s.DeleteReminder))
	reminderRouter.Method(http.MethodPatch, "/{reminderID:[0-9]+}", updatereminder.New(s.UpdateReminder))
	reminderRouter.Method(
		http.MethodPut,
		"/{reminderID:[0-9]+}/channels",
		updatereminderchannels.New(s.UpdateReminderChannels),
	)

	telegramRouter := chi.NewRouter()
	telegramRouter.Method(
		http.MethodPost,
		fmt.Sprintf("/updates/{bot}/%s", deps.Config.TelegramURLSecret),
		telegram.New(deps.Logger, deps.TelegramBotMessageSender, s.VerifyTelegramChannel),
	)

	router := chi.NewRouter()
	router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   deps.Config.AllowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: false,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}))
	router.Use(captcha.SetCaptchaTokenToContext)
	router.Mount("/auth", authRouter)
	router.Mount("/profile", profileRouter)
	router.Mount("/channels", channelsRouter)
	router.Mount("/reminders", reminderRouter)
	router.Mount("/telegram", telegramRouter)

	address := fmt.Sprintf("0.0.0.0:%d", deps.Config.Port)

	return &http.Server{
		Handler: router,
		Addr:    address,
	}
}
