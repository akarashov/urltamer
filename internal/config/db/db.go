package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func Connect(dataBaseDSN string) bool{
    ps := fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=disable", dataBaseDSN, `admin`, `admin`, `demo`)
    db, err := sql.Open("pgx", ps)
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