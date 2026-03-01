package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/akarashov/urltamer/internal/model"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"
)

// PostgresRepository is a repository that stores URL mappings in a PostgreSQL database.
type PostgresRepository struct {
	db *sql.DB
}

// NewPostgresRepository creates a new PostgresRepository.
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

	err = makeMigraton("file://migrations", dataBaseDSN) // migrate
	if err != nil {
		return &PostgresRepository{db: db}, err
	}

	return &PostgresRepository{db: db}, nil
}

// LoadTamers loads all URL mappings from the repository.
func (p *PostgresRepository) LoadTamers(ctx context.Context) (model.Tamers, error) {
	rows, err := p.db.QueryContext(ctx, "SELECT uuid, short_url, original_url, user_id, is_deleted FROM tamers ORDER BY uuid")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var tamers model.Tamers
	for rows.Next() {
		var tamer model.Tamer
		err = rows.Scan(&tamer.ID, &tamer.ShortURL, &tamer.OriginalURL, &tamer.UserID, &tamer.DeletedFlag)
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

// GetUserURLs retrieves all URL mappings for a specific user.
func (p *PostgresRepository) GetUserURLs(ctx context.Context, userID int) (model.Tamers, error) {
	rows, err := p.db.QueryContext(ctx, "SELECT uuid, short_url, original_url, user_id, is_deleted FROM tamers where user_id=$1 ORDER BY uuid", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var tamers model.Tamers
	for rows.Next() {
		var tamer model.Tamer
		err = rows.Scan(&tamer.ID, &tamer.ShortURL, &tamer.OriginalURL, &tamer.UserID, &tamer.DeletedFlag)
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

// InsertTamer inserts a new URL mapping into the repository.
func (p *PostgresRepository) InsertTamer(ctx context.Context, tamer model.Tamer) (int64, error) {
	result, err := p.db.ExecContext(ctx,
		"INSERT INTO tamers (short_url, original_url, user_id) VALUES ($1, $2, $3) ON CONFLICT (original_url, user_id) DO UPDATE SET short_url = $1, user_id = $3",
		tamer.ShortURL, tamer.OriginalURL, tamer.UserID)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

// DeleteTamer marks URL mappings as deleted for a specific user.
func (p *PostgresRepository) DeleteTamer(ctx context.Context, userID int, shortURLs []string) (int64, error) {
	result, err := p.db.ExecContext(ctx,
		"UPDATE tamers SET is_deleted = True WHERE user_id = $1 AND short_url = ANY($2::text[])", userID, shortURLs)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

// CreateUser creates a new user in the repository.
func (p *PostgresRepository) CreateUser(ctx context.Context, userID int) (int64, error) {
	result, err := p.db.ExecContext(ctx,
		"INSERT INTO users (id) VALUES ($1)", userID)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

// GetTamerByShortURL retrieves a URL mapping by its short URL.
func (p *PostgresRepository) GetTamerByShortURL(ctx context.Context, shortURL string) (*model.Tamer, error) {
	var tamer model.Tamer
	err := p.db.QueryRowContext(ctx,
		"SELECT uuid, short_url, original_url, user_id, is_deleted FROM tamers WHERE short_url = $1",
		shortURL,
	).Scan(&tamer.ID, &tamer.ShortURL, &tamer.OriginalURL, &tamer.UserID, &tamer.DeletedFlag)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &tamer, nil
}

// GetTamerByOriginalURL retrieves a URL mapping by its original URL.
func (p *PostgresRepository) GetTamerByOriginalURL(ctx context.Context, originalURL string) (*model.Tamer, error) {
	var tamer model.Tamer
	err := p.db.QueryRowContext(ctx,
		"SELECT uuid, short_url, original_url, user_id, is_deleted FROM tamers WHERE original_url = $1",
		originalURL,
	).Scan(&tamer.ID, &tamer.ShortURL, &tamer.OriginalURL, &tamer.UserID, &tamer.DeletedFlag)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &tamer, nil
}

// Ping checks the connectivity of the repository.
func (p *PostgresRepository) Ping(ctx context.Context) bool {
	err := p.db.PingContext(ctx)
	return err == nil
}

// Close closes the repository.
func (p *PostgresRepository) Close() error {
	return p.db.Close()
}

// makeMigraton applies database migrations.
func makeMigraton(pathMigrations string, dataBaseDSN string) error {
	m, err := migrate.New(pathMigrations, dataBaseDSN)
	if err != nil {
		return err
	}
	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		return err
	}
	return nil
}

// GetUserURLs retrieves all URL mappings for a specific user.
func (p *PostgresRepository) GetInternalStats(ctx context.Context) (*model.InternalStats, error) {
	var internalStats model.InternalStats
	err := p.db.QueryRowContext(ctx, "SELECT count(*) FROM tamers").Scan(&internalStats.URLs)
	if err != nil {
		return nil, err
	}
	err = p.db.QueryRowContext(ctx, "SELECT count(*) FROM users").Scan(&internalStats.Users)
	if err != nil {
		return nil, err
	}
	return &internalStats, nil
}
