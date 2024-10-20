package postgre

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"vieo/auth/internal/domain/models"
	"vieo/auth/internal/storage"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
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
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			if pqErr.Code == "23505" {
				return -1, fmt.Errorf("%s: %w", op, storage.ErrUserAlreadyExists)
			}
		}
		return -1, fmt.Errorf("%s: %w", op, err)
	}

	return lastInsertIndex, nil
}
func (s *Storage) User(
	ctx context.Context,
	email string,
) (models.User, error) {
	const op = "storage.postgres.User"

	var user models.User
	err := s.db.GetContext(ctx, &user, "SELECT * FROM users WHERE email = $1", email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.User{}, fmt.Errorf("%s: user not found: %w", op, storage.ErrUserNotFound)
		}

		return models.User{}, fmt.Errorf("%s: %w", op, err)
	}

	return user, nil

}

func (s *Storage) SaveDevice(
	ctx context.Context,
	email string,
	device string,
) error {
	const op = "storage.postgres.SaveDevice"

	_, err := s.db.ExecContext(ctx, "INSERT INTO devices (email, device_name) VALUES ($1, $2)", email, device)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			switch pqErr.Code {
			case "23505":
				// the user logged in from another browser, the access token is not there, we just generate a new one for him
				return nil
			case "23503":
				return fmt.Errorf("%s: user not found for email: %w", op, storage.ErrUserNotFound)
			case "P0001":
				if pqErr.Message == fmt.Sprintf("Exceeded limit of 5 devices for the same email: %s", email) {
					return fmt.Errorf("%s: %w", op, storage.ErrDeviceLimitExceeded)
				}
			}
		}

		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Storage) Device(
	ctx context.Context,
	email string,
	device string,
) error {
	const op = "storage.postgres.Device"

	var exists bool
	err := s.db.QueryRowContext(ctx,
		"SELECT EXISTS (SELECT 1 FROM devices WHERE email = $1 AND device_name = $2)",
		email,
		device,
	).Scan(&exists)
	if err != nil {
		return fmt.Errorf("%s: failed to execute query: %w", op, err)
	}

	if !exists {
		return fmt.Errorf("%s: %w", op, storage.ErrDeviceNotFound)
	}

	return nil
}
