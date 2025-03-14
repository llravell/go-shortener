package interceptors

import (
	"context"

	"github.com/google/uuid"
	"github.com/llravell/go-shortener/internal/entity"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// Auth предоставляет интерсепторы для работы с авторизацией.
type Auth struct {
	secret           []byte
	protectedMethods map[string]bool
	log              *zerolog.Logger
}

// NewAuth конфигурирует интерсепторы авторизации.
func NewAuth(secretKey string, protectedMethods map[string]bool, log *zerolog.Logger) *Auth {
	return &Auth{
		secret:           []byte(secretKey),
		protectedMethods: protectedMethods,
		log:              log,
	}
}

func (auth *Auth) getTokenFromContext(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ""
	}

	values := md.Get("token")
	if len(values) == 0 {
		return ""
	}

	return values[0]
}

func (auth *Auth) getUserUUIDFromContext(ctx context.Context) string {
	token := auth.getTokenFromContext(ctx)
	if len(token) == 0 {
		return ""
	}

	claims, err := entity.ParseJWTString(token, auth.secret)
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

func (auth *Auth) ProvideJWTInterceptor(
	ctx context.Context,
	req any,
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (any, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		md = metadata.MD{}
	}

	userUUID := auth.getUserUUIDFromContext(ctx)
	if len(userUUID) > 0 {
		md.Append("userUUID", userUUID)

		return handler(metadata.NewIncomingContext(ctx, md), req)
	}

	userUUID = uuid.New().String()

	jwtToken, err := entity.BuildJWTString(userUUID, auth.secret)
	if err != nil {
		auth.log.Error().Err(err).Msg("jwt building failed")

		return handler(ctx, req)
	}

	grpc.SetTrailer(ctx, metadata.Pairs("token", jwtToken))
	md.Append("userUUID", userUUID)

	return handler(metadata.NewIncomingContext(ctx, md), req)
}

func (auth *Auth) CheckJWTInterceptor(
	ctx context.Context,
	req any,
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (any, error) {
	if !auth.protectedMethods[info.FullMethod] {
		return handler(ctx, req)
	}

	userUUID := auth.getUserUUIDFromContext(ctx)
	if len(userUUID) == 0 {
		return nil, status.Error(codes.Unauthenticated, "invalid auth token")
	}

	return handler(ctx, req)
}
