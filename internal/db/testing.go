package db

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/jackc/pgx/v4/pgxpool"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func applyMigrations(connString string) {
	migrationsPath := os.Getenv("TEST_MIGRATIONS_PATH")
	if migrationsPath == "" {
		panic("TEST_MIGRATIONS_PATH must be set.")
	}
	m, err := migrate.New("file://"+migrationsPath, connString)
	if err != nil {
		panic("Could not connect to DB for applying migrations.")
	}
	err = m.Up()
	if !errors.Is(err, migrate.ErrNoChange) && err != nil {
		panic(fmt.Errorf("could not apply DB migrations %w", err))
	}
}

func CreateTestPool() *pgxpool.Pool {
	connString := os.Getenv("TEST_POSTGRESQL_URL")
	if connString == "" {
		panic("TEST_POSTGRESQL_URL must be set.")
	}
	applyMigrations(connString)

	ctx := context.Background()
	pool, err := pgxpool.Connect(ctx, connString)
	if err != nil {
		panic(fmt.Errorf("could not connect to the database %w", err))
	}

	return pool
}

func TruncateTables(pool *pgxpool.Pool) {
	_, err := pool.Exec(context.Background(), "DELETE FROM \"user\";")
	if err != nil {
		panic(fmt.Errorf("could not truncate DB tables %w", err))
	}
}
