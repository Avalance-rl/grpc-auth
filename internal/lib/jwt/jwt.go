package jwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrFailedToExtractData = errors.New("failed to extract data from token")
	ErrTokenExpiration     = errors.New("expiration")
	ErrIncorrectExpiration = errors.New("incorrect token expiration time value")
	ErrInvalidToken        = errors.New("invalid token")
	ErrInvalidSignMethod   = errors.New("invalid signature method")
)

// NewToken generate new access token for client,
// he consists of "email", "deviceAddress", "expiration", "iat"
func NewToken(
	user string,
	duration time.Duration,
	secretKey string,
	deviceAddress string,
) (string, error) {
	jwtSecretKey := []byte(secretKey)
	accessPayload := jwt.MapClaims{
		"email":         user,
		"deviceAddress": deviceAddress,
		"exp":           duration, // interceptor will check access token's duration
		"iat":           time.Now().Unix(),
	}
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessPayload)
	signedAccessToken, err := accessToken.SignedString(jwtSecretKey)
	if err != nil {
		return "", err
	}

	return signedAccessToken, nil
}

// DecodeToken is decoding access token, checking his valid
func DecodeToken(
	secretKey string,
	accessToken string,
) (string, string, error) {
	jwtSecretKey := []byte(secretKey)

	token, err := jwt.Parse(accessToken, func(token *jwt.Token) (interface{}, error) {
		// make sure the correct signature algorithm is used
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidSignMethod
		}
		return jwtSecretKey, nil
	})

	if err != nil {
		return "", "", ErrInvalidSignMethod
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		email, okEmail := claims["email"].(string)
		deviceAddress, okDevice := claims["deviceAddress"].(string)
		if !okEmail || !okDevice {
			return "", "", ErrFailedToExtractData
		}

		expiration, ok := claims["exp"].(float64)
		if !ok {
			return "", "", ErrIncorrectExpiration
		}

		if time.Unix(int64(expiration), 0).Before(time.Now()) {
			return "", "", ErrTokenExpiration
		}

		return email, deviceAddress, nil
	}

	return "", "", ErrInvalidToken
}
