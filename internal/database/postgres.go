package database

import (
	"context"
	"os"

	"github.com/jackc/pgx/v5"
)

func ConnectDB() (*pgx.Conn, error) {
	databaseUrl := os.Getenv("DATABASE_URL")

	conn, err := pgx.Connect(context.Background(), databaseUrl)
	if err != nil {
		return nil, err
	}

	return conn, nil
}
