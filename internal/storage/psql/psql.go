package psql

import (
	"context"
	"database/sql"
	"fmt"
	"time"
	"vieo/auth/internal/domain/models"
)

type Storage interface {
	SaveMovie(ctx context.Context, movie *models.Movie) error
	GetMovie(ctx context.Context, id string) (*models.Movie, error)
	UpdateViewers(ctx context.Context, movieID string, viewers int) error
}

// PostgresStorage реализация хранилища для PostgreSQL
type PostgresStorage struct {
	db *sql.DB
}

func NewPostgresStorage(connStr string) (*PostgresStorage, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to postgres: %w", err)
	}

	// Создаем таблицы если их нет
	_, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS movies (
            id VARCHAR(255) PRIMARY KEY,
            title VARCHAR(255) NOT NULL,
            duration INTEGER NOT NULL,
            current_viewers INTEGER DEFAULT 0,
            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
        )
    `)
	if err != nil {
		return nil, fmt.Errorf("failed to create tables: %w", err)
	}

	return &PostgresStorage{db: db}, nil
}

func (s *PostgresStorage) SaveMovie(ctx context.Context, movie *models.Movie) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO movies (id, title, duration, current_viewers) 
         VALUES ($1, $2, $3, $4)
         ON CONFLICT (id) DO UPDATE 
         SET title = $2, duration = $3, current_viewers = $4`,
		movie.ID, movie.Title, int(movie.Duration.Seconds()), movie.CurrentViewers)
	return err
}

func (s *PostgresStorage) GetMovie(ctx context.Context, id string) (*models.Movie, error) {
	movie := &models.Movie{}
	var durationSec int
	err := s.db.QueryRowContext(ctx,
		"SELECT id, title, duration, current_viewers FROM movies WHERE id = $1",
		id).Scan(&movie.ID, &movie.Title, &durationSec, &movie.CurrentViewers)
	if err != nil {
		return nil, err
	}
	movie.Duration = time.Duration(durationSec) * time.Second
	return movie, nil
}

func (s *PostgresStorage) UpdateViewers(ctx context.Context, movieID string, viewers int) error {
	_, err := s.db.ExecContext(ctx,
		"UPDATE movies SET current_viewers = $1 WHERE id = $2",
		viewers, movieID)
	return err
}
