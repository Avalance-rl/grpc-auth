package auth

import (
	"context"
	"errors"
	"testing"
	"time"
	"vieo/auth/internal/domain/models"
	"vieo/auth/internal/lib/jwt"
	"vieo/auth/internal/lib/logger"
	"vieo/auth/internal/services/auth/mocks"
	"vieo/auth/internal/storage"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

func TestRegisterNewUser(t *testing.T) {
	// Initialize mocks
	mockUserSaver := new(mocks.UserSaver)
	mockLogger := logger.NewLogger("prod")
	authService := New(mockLogger, mockUserSaver, nil, nil, nil, time.Minute, "secretKey")

	// Define test parameters
	email := "test@example.com"
	password := "password123"

	// Success case: user registered successfully
	t.Run("successful registration", func(t *testing.T) {
		mockUserSaver.On("SaveUser", mock.Anything, email, mock.Anything).Return(int64(1), nil).Once()

		// Call RegisterNewUser
		userID, err := authService.RegisterNewUser(context.Background(), email, password)
		// Assertions
		assert.NoError(t, err)
		assert.Equal(t, int64(1), userID)
		mockUserSaver.AssertExpectations(t)
	})

	// Error case: user already exists
	t.Run("user already exists", func(t *testing.T) {
		mockUserSaver.On("SaveUser", mock.Anything, email, mock.Anything).
			Return(int64(-1), storage.ErrUserAlreadyExists).Once()

		// Call RegisterNewUser
		userID, err := authService.RegisterNewUser(context.Background(), email, password)

		// Assertions
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "user already exists")
		assert.Equal(t, int64(-1), userID)
		mockUserSaver.AssertExpectations(t)
	})

	// Error case: failed to hash password
	t.Run("failed to hash password", func(t *testing.T) {
		mockUserSaver.On("SaveUser", mock.Anything, email, mock.Anything).Return(int64(0), errors.New("hash error")).Once()

		// Call RegisterNewUser with a password that will fail the hashing
		userID, err := authService.RegisterNewUser(context.Background(), email, "")

		// Assertions
		assert.Error(t, err)
		assert.Equal(t, int64(0), userID)
		mockUserSaver.AssertExpectations(t)
	})
}

func TestLogin(t *testing.T) {
	// Initialize mocks
	mockUserProvider := new(mocks.UserProvider)
	mockDeviceSaver := new(mocks.DeviceSaver)
	mockLogger := logger.NewLogger("prod")
	authService := New(mockLogger, nil, mockUserProvider, mockDeviceSaver, nil, time.Minute, "secretKey")

	// Define test parameters
	email := "test@example.com"
	password := "password123"
	deviceAddress := "device_1"
	passHash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	user := models.User{Email: email, PassHash: passHash}

	// Success case: successful login
	t.Run("successful login", func(t *testing.T) {
		mockUserProvider.On("User", mock.Anything, email).Return(user, nil).Once()
		mockDeviceSaver.On("SaveDevice", mock.Anything, email, deviceAddress).Return(nil).Once()
		token, _ := jwt.NewToken(email, time.Minute, "secretKey", deviceAddress)

		// Call Login
		resultToken, err := authService.Login(context.Background(), email, password, deviceAddress)

		// Assertions
		assert.NoError(t, err)
		assert.Equal(t, token, resultToken)
		mockUserProvider.AssertExpectations(t)
		mockDeviceSaver.AssertExpectations(t)
	})

	// Error case: user not found
	t.Run("user not found", func(t *testing.T) {
		mockUserProvider.On("User", mock.Anything, email).Return(models.User{}, storage.ErrUserNotFound).Once()

		// Call Login
		_, err := authService.Login(context.Background(), email, password, deviceAddress)

		// Assertions
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid credentials")
		mockUserProvider.AssertExpectations(t)
	})

	// Error case: wrong password
	t.Run("wrong password", func(t *testing.T) {
		incorrectPassword := "wrong_password"
		incorrectPassHash, _ := bcrypt.GenerateFromPassword([]byte(incorrectPassword), bcrypt.DefaultCost)
		user := models.User{Email: email, PassHash: incorrectPassHash}
		mockUserProvider.On("User", mock.Anything, email).Return(user, nil).Once()

		// Call Login
		_, err := authService.Login(context.Background(), email, password, deviceAddress)

		// Assertions
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "wrong password")
		mockUserProvider.AssertExpectations(t)
	})

	// Error case: device limit exceeded
	t.Run("device limit exceeded", func(t *testing.T) {
		mockUserProvider.On("User", mock.Anything, email).Return(user, nil).Once()
		mockDeviceSaver.On("SaveDevice", mock.Anything, email, deviceAddress).Return(storage.ErrDeviceLimitExceeded).Once()

		// Call Login
		_, err := authService.Login(context.Background(), email, password, deviceAddress)

		// Assertions
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "device limit exceeded")
		mockUserProvider.AssertExpectations(t)
		mockDeviceSaver.AssertExpectations(t)
	})
}

func TestRefreshToken(t *testing.T) {
	// Initialize mocks
	mockDeviceProvider := new(mocks.DeviceProvider)
	mockLogger := logger.NewLogger("prod")
	authService := New(mockLogger, nil, nil, nil, mockDeviceProvider, time.Minute, "secretKey")

	// Define test parameters
	email := "test@example.com"
	deviceAddress := "device_1"
	token, _ := jwt.NewToken(email, time.Minute, "secretKey", deviceAddress)

	// Success case: token refreshed successfully
	t.Run("successful token refresh", func(t *testing.T) {
		mockDeviceProvider.On("Device", mock.Anything, email, deviceAddress).Return(nil).Once()

		// Call RefreshToken
		newToken, err := authService.RefreshToken(context.Background(), deviceAddress, token)

		// Assertions
		assert.NoError(t, err)
		assert.NotEmpty(t, newToken)
		mockDeviceProvider.AssertExpectations(t)
	})

	t.Run("address mismatch", func(t *testing.T) {
		wrongDeviceAddress := "wrong_device"

		// Call RefreshToken with a mismatched device address
		_, err := authService.RefreshToken(context.Background(), wrongDeviceAddress, token)

		// Assertions
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "address mismatch")
	})

	// Error case: device not found
	t.Run("device not found", func(t *testing.T) {
		mockDeviceProvider.On("Device", mock.Anything, email, deviceAddress).Return(storage.ErrDeviceNotFound).Once()

		// Call RefreshToken
		_, err := authService.RefreshToken(context.Background(), deviceAddress, token)

		// Assertions
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "device not found")
		mockDeviceProvider.AssertExpectations(t)
	})
}
