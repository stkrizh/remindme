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
	me "remindme/internal/http/handlers/auth/me"
	resetpassword "remindme/internal/http/handlers/auth/reset_password"
	sendpasswordresettoken "remindme/internal/http/handlers/auth/send_password_reset_token"
	signupanonymously "remindme/internal/http/handlers/auth/sign_up_anonymously"
	signupwithemail "remindme/internal/http/handlers/auth/sign_up_with_email"
	createemailchannel "remindme/internal/http/handlers/channels/create_email_channel"
	createtlgchannel "remindme/internal/http/handlers/channels/create_telegram_channel"
	listuserchannels "remindme/internal/http/handlers/channels/list_user_channels"
	verifyemailchannel "remindme/internal/http/handlers/channels/verify_email_channel"
	cancelreminder "remindme/internal/http/handlers/reminders/cancel_reminder"
	createreminder "remindme/internal/http/handlers/reminders/create_reminder"
	listuserreminders "remindme/internal/http/handlers/reminders/list_user_reminders"
	updatereminder "remindme/internal/http/handlers/reminders/update_reminder"
	updatereminderchannels "remindme/internal/http/handlers/reminders/update_reminder_channels"
	telegram "remindme/internal/http/handlers/telegram"
	"time"

	"github.com/go-chi/chi/v5"
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
		"/password_reset/send",
		sendpasswordresettoken.New(s.SendPasswordResetToken, isTestMode),
	)
	authRouter.Method(http.MethodPost, "/password_reset", resetpassword.New(s.ResetPassword))

	profileRouter := chi.NewRouter()
	profileRouter.Use(auth.SetAuthTokenToContext)
	profileRouter.Method(http.MethodGet, "/me", me.New(s.GetUserBySessionToken))

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

	reminderRouter := chi.NewRouter()
	reminderRouter.Use(auth.SetAuthTokenToContext)
	reminderRouter.Method(http.MethodPost, "/", createreminder.New(s.CreateReminder))
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
	router.Mount("/auth", authRouter)
	router.Mount("/profile", profileRouter)
	router.Mount("/channels", channelsRouter)
	router.Mount("/reminders", reminderRouter)
	router.Mount("/telegram", telegramRouter)

	address := fmt.Sprintf("0.0.0.0:%d", deps.Config.Port)

	return &http.Server{
		Handler:           router,
		Addr:              address,
		ReadTimeout:       5 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      5 * time.Second,
		IdleTimeout:       5 * time.Second,
	}
}