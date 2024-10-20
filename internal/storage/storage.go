package storage

import "errors"

var (
	ErrUserAlreadyExists   = errors.New("user already exists")
	ErrUserNotFound        = errors.New("user not found")
	ErrDeviceLimitExceeded = errors.New("device limit exceeded")
	ErrDeviceAlreadyExists = errors.New("device already exists")
	ErrDeviceNotFound      = errors.New("device not found")
)
