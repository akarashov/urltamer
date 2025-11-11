package repository

import (
	"database/sql"
	"context"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/akarashov/urltamer/internal/model"
)

func Connect(dataBaseDSN string) (context.Context, *sql.DB, error){
    db, err := sql.Open("pgx", dataBaseDSN)
    if err != nil {
		return nil, nil, err
    }
	defer db.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
    defer cancel()
	if err = db.PingContext(ctx); err != nil {
		return  nil, nil, err
    }
	return ctx, db, nil
}

func SelectTamers(ctx context.Context, db *sql.DB) (model.Tamers, error) {
    var tamers model.Tamers
    rows, err := db.QueryContext(ctx, "SELECT uuid, short_url, original_url, from tamers")
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    for rows.Next() {
        var t model.Tamer
        err = rows.Scan(&t.UUID, &t.ShortURL, &t.OriginalURL)
        if err != nil {
            return nil, err
        }
        tamers = append(tamers, t)
    }
    err = rows.Err()
    if err != nil {
        return nil, err
    }
    return tamers, nil
}

func InsertTamer(tamer model.Tamer, db *sql.DB) (int64, error) {
    result, err := db.Exec("INSERT INTO tamers (short_url, original_url) VALUES ($1, $2)", tamer.ShortURL, tamer.OriginalURL)
    if err != nil {
        return 0, err
    }
    return result.LastInsertId()
}