// internal/repository/postgres.go
package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/akarashov/urltamer/internal/model"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(dataBaseDSN string) (*PostgresRepository, error) {
	db, err := sql.Open("pgx", dataBaseDSN)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err = db.PingContext(ctx); err != nil {
		return nil, err
	}
	return &PostgresRepository{db: db}, nil
}

func (p *PostgresRepository) LoadTamers(ctx context.Context) (model.Tamers, error) {
	rows, err := p.db.QueryContext(ctx, "SELECT uuid, short_url, original_url FROM tamers ORDER BY uuid")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var tamers model.Tamers
	for rows.Next() {
		var tamer model.Tamer
		err = rows.Scan(&tamer.ID, &tamer.ShortURL, &tamer.OriginalURL)
		if err != nil {
			return nil, err
		}
		tamers = append(tamers, tamer)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return tamers, nil
}

func (p *PostgresRepository) InsertTamer(ctx context.Context, tamer model.Tamer) (int64, error) {
	result, err := p.db.ExecContext(ctx,
		"INSERT INTO tamers (short_url, original_url) VALUES ($1, $2) ON CONFLICT (original_url) DO UPDATE SET short_url = $1",
		tamer.ShortURL, tamer.OriginalURL)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func (p *PostgresRepository) GetTamerByShortURL(ctx context.Context, shortURL string) (*model.Tamer, error) {
	var tamer model.Tamer
	err := p.db.QueryRowContext(ctx,
		"SELECT uuid, short_url, original_url FROM tamers WHERE short_url = $1",
		shortURL,
	).Scan(&tamer.ID, &tamer.ShortURL, &tamer.OriginalURL)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &tamer, nil
}

func (p *PostgresRepository) GetTamerByOriginalURL(ctx context.Context, originalURL string) (*model.Tamer, error) {
	var tamer model.Tamer
	err := p.db.QueryRowContext(ctx,
		"SELECT uuid, short_url, original_url FROM tamers WHERE original_url = $1",
		originalURL,
	).Scan(&tamer.ID, &tamer.ShortURL, &tamer.OriginalURL)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &tamer, nil
}

func (p *PostgresRepository) Ping(ctx context.Context) bool {
	err := p.db.PingContext(ctx)
	return err == nil
}

func (p *PostgresRepository) Close() error {
	return p.db.Close()
}
