package entity

import (
	"time"

	"github.com/golang-jwt/jwt/v4"
)

const JWTExpire = time.Hour * 3

type JWTClaims struct {
	jwt.RegisteredClaims
	UserUUID string
}

func BuildJWTString(userUUID string, secret []byte) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, JWTClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(JWTExpire)),
		},
		UserUUID: userUUID,
	})

	tokenString, err := token.SignedString(secret)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func ParseJWTString(tokenString string, secret []byte) (*JWTClaims, error) {
	claims := &JWTClaims{}

	_, err := jwt.ParseWithClaims(tokenString, claims, func(_ *jwt.Token) (interface{}, error) {
		return secret, nil
	})
	if err != nil {
		return nil, err
	}

	return claims, nil
}
