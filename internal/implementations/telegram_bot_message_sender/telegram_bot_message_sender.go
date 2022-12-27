package telegrambotmessagesender

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"remindme/internal/core/domain/bot"
	"remindme/internal/core/domain/channel"
	"time"
)

type telegramMessage struct {
	ChatID int64  `json:"chat_id"`
	Text   string `json:"text"`
}

type TelegramBotMessageSender struct {
	httpClient http.Client
	baseURL    url.URL
	tokenByBot map[channel.TelegramBot]string
}

func New(
	baseURL url.URL,
	tokenByBot map[channel.TelegramBot]string,
	timeout time.Duration,
) *TelegramBotMessageSender {
	return &TelegramBotMessageSender{
		baseURL:    baseURL,
		tokenByBot: tokenByBot,
		httpClient: http.Client{Timeout: timeout},
	}
}

func (s *TelegramBotMessageSender) SendTelegramBotMessage(ctx context.Context, m bot.TelegramBotMessage) error {
	token, ok := s.tokenByBot[m.Bot]
	if !ok {
		return fmt.Errorf("bot token not found, message: %v, tokens: %v", m, s.tokenByBot)
	}
	url := s.baseURL.JoinPath(fmt.Sprintf("bot%s", token), "sendMessage")
	var body bytes.Buffer
	encoder := json.NewEncoder(&body)
	err := encoder.Encode(telegramMessage{ChatID: int64(m.ChatID), Text: m.Text})
	if err != nil {
		return err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, url.String(), &body)
	if err != nil {
		return err
	}
	request.Header.Add("content-type", "application/json")
	resp, err := s.httpClient.Do(request)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		return fmt.Errorf("got unsuccessfull response from Telegram: %s", string(body))
	}
	return nil
}
