package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/llravell/go-shortener/internal/entity"
	"github.com/rs/zerolog"
)

const (
	TokenCookieName = "user-token"
)

type contextKey string

var UserUUIDContextKey contextKey = "userUUID"

type Auth struct {
	secret []byte
	log    *zerolog.Logger
}

func (auth *Auth) parseUserUUIDFromRequest(r *http.Request) string {
	tokenCookie, err := r.Cookie(TokenCookieName)
	if err != nil {
		auth.log.Error().Err(err).Msg("jwt cookie finding failed")

		return ""
	}

	claims, err := entity.ParseJWTString(tokenCookie.Value, auth.secret)
	if err != nil {
		auth.log.Error().Err(err).Msg("jwt parsing failed")

		return ""
	}

	if err = claims.Valid(); err != nil {
		auth.log.Error().Err(err).Msg("got invalid jwt")

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
			auth.log.Error().Err(err).Msg("jwt building failed")
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

		next.ServeHTTP(w, auth.provideUserUUIDToRequestContext(r, userUUID))
	})
}

func NewAuth(secretKey string, log *zerolog.Logger) *Auth {
	auth := &Auth{
		secret: []byte(secretKey),
		log:    log,
	}

	return auth
}
