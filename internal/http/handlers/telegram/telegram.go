package telegram

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"remindme/internal/core/domain/bot"
	"remindme/internal/core/domain/channel"
	e "remindme/internal/core/domain/errors"
	"remindme/internal/core/domain/logging"
	"remindme/internal/core/services"
	channelVerificationService "remindme/internal/core/services/verify_telegram_channel"
	"remindme/internal/http/handlers/response"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
)

const CHANNEL_VERIFICATION_CODE = "vrf"

type Handler struct {
	log                 logging.Logger
	botMessageSender    bot.TelegramBotMessageSender
	channelVerification services.Service[channelVerificationService.Input, channelVerificationService.Result]
}

func New(
	log logging.Logger,
	botMessageSender bot.TelegramBotMessageSender,
	channelVerification services.Service[channelVerificationService.Input, channelVerificationService.Result],
) *Handler {
	if log == nil {
		panic(e.NewNilArgumentError("log"))
	}
	if botMessageSender == nil {
		panic(e.NewNilArgumentError("botMessageSender"))
	}
	if channelVerification == nil {
		panic(e.NewNilArgumentError("channelVerification"))
	}
	return &Handler{
		log:                 log,
		botMessageSender:    botMessageSender,
		channelVerification: channelVerification,
	}
}

type user struct {
	ID int64 `json:"id"`
}

type message struct {
	ID   int64  `json:"message_id"`
	From user   `json:"from"`
	Date int64  `json:"date"`
	Text string `json:"text"`
}

type update struct {
	ID      int64    `json:"update_id"`
	Message *message `json:"message"`
}

func (u *update) FromJSON(r io.Reader) error {
	e := json.NewDecoder(r)
	return e.Decode(u)
}

func (h *Handler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	defer response.Render(rw, struct{}{}, http.StatusOK)

	bot := channel.TelegramBot(chi.URLParam(r, "bot"))

	update := update{}
	if err := update.FromJSON(r.Body); err != nil {
		h.log.Error(
			r.Context(),
			"Could not decode Telegram update.",
			logging.Entry("err", err),
		)
		return
	}
	if update.Message == nil {
		h.log.Info(
			r.Context(),
			"Skip Telegram update.",
			logging.Entry("update", update),
		)
		return
	}
	h.log.Info(
		r.Context(),
		"Got Telegram update.",
		logging.Entry("bot", bot),
		logging.Entry("updateID", update.ID),
		logging.Entry("updateMessage", update.Message),
	)

	verificationData, ok := parseVerificationData(update)
	if !ok {
		h.sendBotMessage(r.Context(), bot, update.Message.From.ID, "Please visit https://remindme.one")
		return
	}
	_, err := h.channelVerification.Run(
		r.Context(),
		channelVerificationService.Input{
			ChannelID:         verificationData.channelID,
			VerificationToken: verificationData.token,
			TelegramBot:       bot,
			TelegramChatID:    verificationData.telegramChatID,
		},
	)
	if err != nil {
		h.sendBotMessage(
			r.Context(),
			bot,
			update.Message.From.ID,
			"Sorry 😔, the verification code is not valid. Please try again later.",
		)
		return
	}

	h.sendBotMessage(
		r.Context(),
		bot,
		update.Message.From.ID,
		"Thank you, verification succeeded. You may visit https://remindme.one and create reminders!",
	)
}

func (h *Handler) sendBotMessage(ctx context.Context, b channel.TelegramBot, chatID int64, text string) {
	err := h.botMessageSender.SendTelegramBotMessage(ctx, bot.TelegramBotMessage{
		Bot:    b,
		ChatID: channel.TelegramChatID(chatID),
		Text:   text,
	})
	if err != nil {
		h.log.Error(
			ctx,
			"Could not send Telegram bot message due to unexpected error.",
			logging.Entry("fromID", chatID),
			logging.Entry("text", text),
			logging.Entry("err", err),
		)
		return
	}
	h.log.Info(
		ctx,
		"Telegram message successfully sent.",
		logging.Entry("bot", b),
		logging.Entry("chatID", chatID),
	)
}

type channelVerificationData struct {
	channelID      channel.ID
	token          channel.VerificationToken
	telegramChatID channel.TelegramChatID
}

func parseVerificationData(u update) (d channelVerificationData, ok bool) {
	message := strings.TrimPrefix(u.Message.Text, "/start ")
	parts := strings.SplitN(message, "-", 3)
	if len(parts) != 3 {
		return d, false
	}
	if parts[0] != CHANNEL_VERIFICATION_CODE {
		return d, false
	}
	channelID, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return d, false
	}
	return channelVerificationData{
		channelID:      channel.ID(channelID),
		token:          channel.VerificationToken(parts[2]),
		telegramChatID: channel.TelegramChatID(u.Message.From.ID),
	}, true
}
