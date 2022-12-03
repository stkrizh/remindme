package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	Secret           string
	PostgresqlURL    string
	BcryptHasherCost int
}

func Load() (*Config, error) {
	postgresqlURL := os.Getenv("POSTGRESQL_URL")
	if postgresqlURL == "" {
		return nil, fmt.Errorf("POSTGRESQL_URL must be set")
	}

	secret := os.Getenv("SECRET")
	if secret == "" {
		return nil, fmt.Errorf("SECRET must be set")
	}

	bcryptHasherCostRaw := os.Getenv("BCRYPT_HASHER_COST")
	bcryptHasherCost, err := strconv.Atoi(bcryptHasherCostRaw)
	if err != nil {
		return nil, fmt.Errorf("invalid BCRYPT_HASHER_COST value: %w", err)
	}

	return &Config{
		Secret:           secret,
		PostgresqlURL:    postgresqlURL,
		BcryptHasherCost: bcryptHasherCost,
	}, nil
}
