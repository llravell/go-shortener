package middleware

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

const (
	tokenExp        = time.Hour * 3
	TokenCookieName = "user-token"
)

type contextKey string

var UserUUIDContextKey contextKey = "userUUID"

type auth struct {
	secret []byte
}

type authClaims struct {
	jwt.RegisteredClaims
	UserUUID string
}

func (a auth) buildJWTString(userUUID string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, authClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(tokenExp)),
		},
		UserUUID: userUUID,
	})

	tokenString, err := token.SignedString(a.secret)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (a auth) parseUserUUIDFromRequest(r *http.Request) string {
	tokenCookie, err := r.Cookie(TokenCookieName)
	if err != nil {
		return ""
	}

	claims := &authClaims{}

	token, err := jwt.ParseWithClaims(tokenCookie.Value, claims, func(_ *jwt.Token) (interface{}, error) {
		return a.secret, nil
	})

	if err != nil || !token.Valid {
		return ""
	}

	return claims.UserUUID
}

func (a auth) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userUUID := a.parseUserUUIDFromRequest(r)
		fmt.Println("---->", userUUID)

		if userUUID == "" {
			userUUID = uuid.New().String()

			jwtToken, err := a.buildJWTString(userUUID)
			if err != nil {
				next.ServeHTTP(w, r)

				return
			}

			http.SetCookie(w, &http.Cookie{
				Name:     TokenCookieName,
				MaxAge:   int(tokenExp.Seconds()),
				HttpOnly: true,
				Value:    jwtToken,
			})
		}

		ctx := context.WithValue(r.Context(), UserUUIDContextKey, userUUID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func AuthMiddleware(secretKey string) func(next http.Handler) http.Handler {
	a := auth{
		secret: []byte(secretKey),
	}

	return a.Handler
}
