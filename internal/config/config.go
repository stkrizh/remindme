package config

import (
	"fmt"
	"net/url"
	"remindme/internal/core/domain/channel"
	"time"

	"github.com/caarlos0/env/v6"
)

type Config struct {
	IsTestMode                      bool          `env:"TEST_MODE" envDefault:"false"`
	BaseURL                         url.URL       `env:"BASE_URL" envDefault:"localhost"`
	Port                            uint16        `env:"PORT" envDefault:"9090"`
	Secret                          string        `env:"SECRET,notEmpty"`
	PostgresqlURL                   string        `env:"POSTGRESQL_URL,notEmpty"`
	RedisURL                        string        `env:"REDIS_URL,notEmpty"`
	BcryptHasherCost                int           `env:"BCRYPT_HASHER_COST" envDefault:"10"`
	PasswordResetValidDurationHours int           `env:"PASSWORD_RESET_VALIDATION_HOURS" envDefault:"24"`
	TelegramURLSecret               string        `env:"TELEGRAM_URL_SECRET,notEmpty"`
	TelegramBaseURL                 url.URL       `env:"TELEGRAM_BASE_URL" envDefault:"https://api.telegram.org"`
	TelegramBots                    []string      `env:"TELEGRAM_BOTS,notEmpty"`
	TelegramTokens                  []string      `env:"TELEGRAM_TOKENS,notEmpty"`
	TelegramRequestTimeout          time.Duration `env:"TELEGRAM_REQUEST_DURATION" envDefault:"30s"`
}

func Load() (*Config, error) {
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return cfg, err
	}
	if len(cfg.TelegramBots) == 0 || len(cfg.TelegramBots) != len(cfg.TelegramTokens) {
		return cfg, fmt.Errorf(
			"invalid telegram bots (%v) or tokens (%v)",
			cfg.TelegramBots,
			cfg.TelegramTokens,
		)
	}
	return cfg, nil
}

func (c *Config) TelegramTokenByBot() map[channel.TelegramBot]string {
	if len(c.TelegramBots) == 0 || len(c.TelegramBots) != len(c.TelegramTokens) {
		panic("invalid telegram bots or tokens")
	}
	m := make(map[channel.TelegramBot]string)
	for i := 0; i < len(c.TelegramBots); i++ {
		m[channel.TelegramBot(c.TelegramBots[i])] = c.TelegramTokens[i]
	}
	return m
}
