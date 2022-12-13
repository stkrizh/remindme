package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	IsTestMode                      bool
	Secret                          string
	PostgresqlURL                   string
	RedisURL                        string
	BcryptHasherCost                int
	PasswordResetValidDurationHours int
}

func Load() (*Config, error) {
	isTestMode := os.Getenv("TEST_MODE") == "true"

	secret := os.Getenv("SECRET")
	if secret == "" {
		return nil, fmt.Errorf("SECRET must be set")
	}

	postgresqlURL := os.Getenv("POSTGRESQL_URL")
	if postgresqlURL == "" {
		return nil, fmt.Errorf("POSTGRESQL_URL must be set")
	}

	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		return nil, fmt.Errorf("REDIS_URL must be set")
	}

	bcryptHasherCostRaw := os.Getenv("BCRYPT_HASHER_COST")
	bcryptHasherCost, err := strconv.Atoi(bcryptHasherCostRaw)
	if err != nil {
		return nil, fmt.Errorf("invalid BCRYPT_HASHER_COST value: %w", err)
	}

	passwordResetValidDurationHoursRaw := os.Getenv("PASSWORD_RESET_VALID_DURATION_HOURS")
	passwordResetValidDurationHours, err := strconv.Atoi(passwordResetValidDurationHoursRaw)
	if err != nil {
		return nil, fmt.Errorf("invalid PASSWORD_RESET_VALID_DURATION_HOURS value: %w", err)
	}

	return &Config{
		IsTestMode:                      isTestMode,
		Secret:                          secret,
		PostgresqlURL:                   postgresqlURL,
		RedisURL:                        redisURL,
		BcryptHasherCost:                bcryptHasherCost,
		PasswordResetValidDurationHours: passwordResetValidDurationHours,
	}, nil
}
