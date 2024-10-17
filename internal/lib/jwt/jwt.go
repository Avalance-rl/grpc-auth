package jwt

import (
	"time"
	"vieo/auth/internal/domain/models"

	"github.com/golang-jwt/jwt/v5"
)

const (
	refreshLife = 3600
)

// TODO: не забыть переделать секрет токена, покрыть функцию тестами

func NewToken(
	user models.User,
	duration time.Duration,
	secretKey string,
	deviceAddress string,
) (string, error) {
	jwtSecretKey := []byte(secretKey)
	accessPayload := jwt.MapClaims{
		"email":         user.Email,
		"deviceAddress": deviceAddress,
		"exp":           duration,
	}
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessPayload)
	signedAccessToken, err := accessToken.SignedString(jwtSecretKey)
	if err != nil {
		return "", err
	}

	return signedAccessToken, nil

}
