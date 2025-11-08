package db

import (
	"context"
	"database/sql"
	// "fmt"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func Connect(dataBaseDSN string) bool{
    db, err := sql.Open("pgx", dataBaseDSN)
    if err != nil {
		return false
    }
	defer db.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
    defer cancel()
	if err = db.PingContext(ctx); err != nil {
		return false
    }
	return true
}