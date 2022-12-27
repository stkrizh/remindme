package main

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"remindme/internal/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	for bot, token := range cfg.TelegramTokenByBot() {
		url := cfg.BaseURL.JoinPath("telegram", "updates", string(bot), cfg.TelegramURLSecret)
		buf := bytes.NewBufferString(fmt.Sprintf(`{"url": "%s"`, url))
		resp, err := http.Post(
			cfg.TelegramBaseURL.JoinPath(fmt.Sprintf("bot%s", token), "setWebhook").String(),
			"application/json",
			buf,
		)
		if err != nil {
			fmt.Fprintf(os.Stderr, "could not register telegram webhook for bot %s, error: %v\n", bot, err)
			os.Exit(1)
		}
		if resp.StatusCode != http.StatusOK {
			fmt.Fprintf(os.Stderr, "could not register telegram webhook for bot %s, status: %v\n", bot, resp.StatusCode)
			os.Exit(1)
		}

		fmt.Printf("Webhook %s successfully registered for bot %s\n", url, bot)
	}
}
