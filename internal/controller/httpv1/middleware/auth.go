package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/llravell/go-shortener/internal/entity"
)

const (
	TokenCookieName = "user-token"
)

type contextKey string

var UserUUIDContextKey contextKey = "userUUID"

type Auth struct {
	secret []byte
}

func (auth *Auth) parseUserUUIDFromRequest(r *http.Request) string {
	tokenCookie, err := r.Cookie(TokenCookieName)
	if err != nil {
		return ""
	}

	claims, err := entity.ParseJWTString(tokenCookie.Value, auth.secret)
	if err != nil || claims.Valid() != nil {
		return ""
	}

	return claims.UserUUID
}

func (auth *Auth) provideUserUUIDToRequestContext(r *http.Request, userUUID string) *http.Request {
	ctx := context.WithValue(r.Context(), UserUUIDContextKey, userUUID)

	return r.WithContext(ctx)
}

func (auth *Auth) ProvideJWTMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userUUID := auth.parseUserUUIDFromRequest(r)

		if userUUID != "" {
			next.ServeHTTP(w, auth.provideUserUUIDToRequestContext(r, userUUID))

			return
		}

		userUUID = uuid.New().String()

		jwtToken, err := entity.BuildJWTString(userUUID, auth.secret)
		if err != nil {
			next.ServeHTTP(w, r)

			return
		}

		http.SetCookie(w, &http.Cookie{
			Name:     TokenCookieName,
			MaxAge:   int(entity.JWTExpire.Seconds()),
			HttpOnly: true,
			Value:    jwtToken,
		})

		next.ServeHTTP(w, auth.provideUserUUIDToRequestContext(r, userUUID))
	})
}

func (auth *Auth) CheckJWTMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userUUID := auth.parseUserUUIDFromRequest(r)

		if userUUID == "" {
			w.WriteHeader(http.StatusUnauthorized)

			return
		}

		ctx := context.WithValue(r.Context(), UserUUIDContextKey, userUUID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func NewAuth(secretKey string) *Auth {
	auth := &Auth{
		secret: []byte(secretKey),
	}

	return auth
}
