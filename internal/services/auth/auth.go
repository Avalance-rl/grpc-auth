package auth

import (
	"context"
	"errors"
	"fmt"
	"time"
	"vieo/auth/internal/domain/models"
	"vieo/auth/internal/lib/jwt"
	"vieo/auth/internal/lib/logger"
	"vieo/auth/internal/storage"

	_ "github.com/lib/pq"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

// TODO: подумать про кэш. после написания основного сервиса

type Auth struct {
	log         *logger.Logger
	usrSaver    UserSaver
	usrProvider UserProvider
	deviceSaver DeviceSaver
	tokenTTL    time.Duration
	secretKey   string
}

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
)

type UserSaver interface {
	SaveUser(
		ctx context.Context,
		email string,
		passHash []byte,
	) (uid int64, err error)
}

type UserProvider interface {
	User(
		ctx context.Context,
		email string,
	) (models.User, error)
}

type DeviceSaver interface {
	SaveDevice(
		ctx context.Context,
		email string,
		device string,
	) error
}

func New(
	log *logger.Logger,
	userSaver UserSaver,
	userProvider UserProvider,
	deviceSaver DeviceSaver,
	tokenTTL time.Duration,
	secretKey string,
) *Auth {
	return &Auth{
		usrSaver:    userSaver,
		usrProvider: userProvider,
		deviceSaver: deviceSaver,
		log:         log,
		tokenTTL:    tokenTTL,
		secretKey:   secretKey,
	}
}

func (a *Auth) RegisterNewUser(
	ctx context.Context,
	email string,
	password string,
) (userID int64, err error) {
	const op = "Auth.RegisterNewUser"
	log := a.log.With(
		zap.String("op", op),
		zap.String("email", "****"+email[4:]),
	)
	log.Info("registering user")

	passHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Error("failed to hash password", zap.Error(err))

		return 0, fmt.Errorf("%s: %w", op, err)
	}
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	id, err := a.usrSaver.SaveUser(ctx, email, passHash)
	if err != nil {
		log.Error("failed to save user", zap.Error(err))

		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return id, nil
}

// TODO: с 2 jwt переписать

func (a *Auth) Login(
	ctx context.Context,
	email string,
	password string,
	deviceAddress string,
) (string, error) {
	const op = "Auth.Login"

	log := a.log.With(
		zap.String("op", op),
		zap.String("email", email),
	)
	log.Info("attempting to login user")

	user, err := a.usrProvider.User(ctx, email)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			a.log.Warn("user not found", zap.Error(err))

			return "", fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
		}

		a.log.Error("failed to get user", zap.Error(err))

		return "", fmt.Errorf("%s: %w", op, err)
	}
	if err := bcrypt.CompareHashAndPassword(user.PassHash, []byte(password)); err != nil {
		a.log.Info("invalid credentials", zap.Error(err))
		return "", fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
	}
	log.Info("successfully logged in")
	err = a.deviceSaver.SaveDevice(ctx, email, "")
	if err != nil {
		// TODO: по сути либо много устройств, либо с бд проблемы
	}
	token, err := jwt.NewToken(user, a.tokenTTL, a.secretKey, "")
	if err != nil {
		a.log.Error("failed to generate token", zap.Error(err))

		return "", fmt.Errorf("%s: %w", op, err)
	}

	return token, nil
}

func (a *Auth) RefreshToken(
	ctx context.Context,
	deviceAddress string,
) (string, error) {

	return "", nil
}
