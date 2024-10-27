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

	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

//go:generate mockery --case=underscore --all

type Auth struct {
	log            *logger.Logger
	usrSaver       UserSaver
	usrProvider    UserProvider
	deviceSaver    DeviceSaver
	deviceProvider DeviceProvider
	tokenTTL       time.Duration
	secretKey      string
}

const (
	queryTime = 3 * time.Second
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrWrongPassword      = errors.New("wrong password")
	ErrAddressMismatch    = errors.New("address mismatch")
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

type DeviceProvider interface {
	Device(
		ctx context.Context,
		device string,
		email string,
	) error
}

func New(
	log *logger.Logger,
	userSaver UserSaver,
	userProvider UserProvider,
	deviceSaver DeviceSaver,
	deviceProvider DeviceProvider,
	tokenTTL time.Duration,
	secretKey string,
) *Auth {
	return &Auth{
		usrSaver:       userSaver,
		usrProvider:    userProvider,
		deviceSaver:    deviceSaver,
		deviceProvider: deviceProvider,
		log:            log,
		tokenTTL:       tokenTTL,
		secretKey:      secretKey,
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
	ctx, cancel := context.WithTimeout(ctx, queryTime)
	defer cancel()
	id, err := a.usrSaver.SaveUser(ctx, email, passHash)
	if err != nil {
		if errors.Is(err, storage.ErrUserAlreadyExists) {
			a.log.Warn("user already exists", zap.Error(err))
			return -1, fmt.Errorf("%s: %w", op, err)
		}
		log.Error("failed to save user", zap.Error(err))

		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return id, nil
}

func (a *Auth) Login(
	ctx context.Context,
	email string,
	password string,
	deviceAddress string,
) (string, error) {
	const op = "Auth.Login"

	log := a.log.With(
		zap.String("op", op),
		zap.String("email", "****"+email[4:]),
	)
	log.Info("attempting to login user")
	ctx, cancel := context.WithTimeout(ctx, queryTime)
	defer cancel()
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
		return "", fmt.Errorf("%s: %w", op, ErrWrongPassword)
	}
	log.Info("successfully logged in")
	ctx, cancel = context.WithTimeout(ctx, queryTime)
	defer cancel()
	err = a.deviceSaver.SaveDevice(ctx, email, deviceAddress)
	if err != nil {
		if errors.Is(err, storage.ErrDeviceLimitExceeded) {
			a.log.Warn("device limit exceeded", zap.Error(err))
			return "", fmt.Errorf("%s: %w", op, err)
		}
		if errors.Is(err, storage.ErrUserNotFound) {
			a.log.Warn("user not found", zap.Error(err))
			return "", fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
		}
		a.log.Error("failed to save device", zap.Error(err))
		return "", fmt.Errorf("%s: %w", op, err)
	}
	token, err := jwt.NewToken(user.Email, a.tokenTTL, a.secretKey, deviceAddress)
	if err != nil {
		a.log.Error("failed to generate token", zap.Error(err))

		return "", fmt.Errorf("%s: %w", op, err)
	}

	return token, nil
}

func (a *Auth) RefreshToken(
	ctx context.Context,
	deviceAddress string,
	accessToken string,
) (string, error) {
	const op = "Auth.Login"

	log := a.log.With(
		zap.String("op", op),
		// hide address and token
		zap.String("deviceAddress", deviceAddress),
		zap.String("accessToken", accessToken),
	)
	log.Info("attempting to refresh token")
	ctx, cancel := context.WithTimeout(ctx, queryTime)
	defer cancel()

	email, addressInToken, err := jwt.DecodeToken(a.secretKey, accessToken)
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpiration) {
			a.log.Warn("token expired", zap.Error(err))
			return "", fmt.Errorf("%s: %w", op, err)
		}
		if errors.Is(err, jwt.ErrInvalidSignMethod) {
			a.log.Warn("invalid sign method", zap.Error(err))
			return "", fmt.Errorf("%s: %w", op, err)
		}
		if errors.Is(err, jwt.ErrFailedToExtractData) {
			a.log.Warn("failed to extract data", zap.Error(err))
			return "", fmt.Errorf("%s: %w", op, err)
		}
		if errors.Is(err, jwt.ErrIncorrectExpiration) {
			a.log.Warn("incorrect expiration", zap.Error(err))
			return "", fmt.Errorf("%s: %w", op, err)
		}
		if errors.Is(err, jwt.ErrInvalidToken) {
			a.log.Warn("invalid token", zap.Error(err))
			return "", fmt.Errorf("%s: %w", op, err)
		}
		a.log.Error("failed to decode token", zap.Error(err))
		return "", fmt.Errorf("%s: %w", op, err)
	}

	// check is needed so that there is no possibility
	// use someone else's token. Explicitly checking devices
	if addressInToken != deviceAddress {
		a.log.Warn("address is mismatch")
		return "", ErrAddressMismatch
	}
	ctx, cancel = context.WithTimeout(ctx, queryTime)
	defer cancel()

	err = a.deviceProvider.Device(ctx, email, deviceAddress)
	if err != nil {
		if errors.Is(err, storage.ErrDeviceNotFound) {
			a.log.Warn("device not found", zap.Error(err))
			return "", fmt.Errorf("%s: %w", op, err)
		}
		a.log.Error("failed to check device", zap.Error(err))
		return "", fmt.Errorf("%s: %w", op, err)
	}

	token, err := jwt.NewToken(email, a.tokenTTL, a.secretKey, deviceAddress)
	if err != nil {
		a.log.Error("failed to generate token", zap.Error(err))
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return token, nil
}
