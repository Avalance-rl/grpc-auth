package postgre

import (
	"context"
	"fmt"
	"vieo/auth/internal/domain/models"

	"github.com/jackc/pgx"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type Storage struct {
	db *sqlx.DB
}

func New(storagePath string) (*Storage, error) {
	const op = "storage.postgres.New"
	db, err := sqlx.Open("postgres", storagePath)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	db.MustExec(models.Schema)
	return &Storage{db: db}, nil
}

func (s *Storage) SaveUser(
	ctx context.Context,
	email string,
	passHash []byte,
) (int64, error) {
	const op = "storage.postgres.SaveUser"

	var lastInsertIndex int64

	err := s.db.QueryRowContext(
		ctx,
		"INSERT INTO users (email, password) VALUES ($1, $2) RETURNING id",
		email,
		passHash,
	).Scan(&lastInsertIndex)

	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return lastInsertIndex, nil
}
func (s *Storage) User(
	ctx context.Context,
	email string,
) (models.User, error) {
	const op = "storage.postgres.User"
	// TODO: дописать запрос

	var user models.User
	err := s.db.GetContext(ctx, &user, "SELECT * FROM users WHERE email = $1", email)
	if err != nil {
		var postgresErr pgx.PgError
		// TODO: обработать разные ошибки
		_ = postgresErr
		return models.User{}, err
	}

	return user, nil

}

func (s *Storage) SaveDevice(
	ctx context.Context,
	email string,
	device string,
) error {
	panic("implement me")
	return nil
}
