package config

import (
	"github.com/caarlos0/env/v6"
)

type Config struct {
	IsTestMode                      bool   `env:"TEST_MODE" envDefault:"false"`
	Secret                          string `env:"SECRET,notEmpty"`
	PostgresqlURL                   string `env:"POSTGRESQL_URL,notEmpty"`
	RedisURL                        string `env:"REDIS_URL,notEmpty"`
	BcryptHasherCost                int    `env:"BCRYPT_HASHER_COST" envDefault:"10"`
	PasswordResetValidDurationHours int    `env:"PASSWORD_RESET_VALIDATION_HOURS" envDefault:"24"`
}

func Load() (*Config, error) {
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return cfg, err
	}
	return cfg, nil
}
