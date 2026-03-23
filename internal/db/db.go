package db

import (
	"context"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

var DB *pgxpool.Pool

func Init() error {
	var err error
	DB, err = pgxpool.New(context.Background(), os.Getenv("DATABASE_URL"))
	return err
}
