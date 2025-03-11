package server

import (
	"context"
	"errors"

	"github.com/llravell/go-shortener/internal/entity"
	pb "github.com/llravell/go-shortener/internal/proto"
	"github.com/llravell/go-shortener/internal/usecase"
	"github.com/rs/zerolog"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// URLUseCase юзкейс базовых операций с урлами.
type URLUseCase interface {
	SaveURL(ctx context.Context, url string, userUUID string) (*entity.URL, error)
	SaveURLMultiple(ctx context.Context, urls []string, userUUID string) ([]*entity.URL, error)
	ResolveURL(ctx context.Context, hash string) (*entity.URL, error)
	GetUserURLS(ctx context.Context, userUUID string) ([]*entity.URL, error)
	BuildRedirectURL(url *entity.URL) string
	QueueDelete(item *entity.URLDeleteItem) error
}

type ShortenerServer struct {
	pb.UnimplementedShortenerServer

	urlUC URLUseCase
	log   *zerolog.Logger
}

func NewShortenerServer(
	urlUC URLUseCase,
	log *zerolog.Logger,
) *ShortenerServer {
	return &ShortenerServer{
		urlUC: urlUC,
		log:   log,
	}
}

func (s *ShortenerServer) ShortenURL(ctx context.Context, in *pb.ShortenURLRequest) (*pb.ShortenURLResponse, error) {
	if len(in.Url) == 0 {
		return nil, status.Error(codes.InvalidArgument, "invalid url")
	}

	url, err := s.urlUC.SaveURL(ctx, in.Url, "")
	if err != nil {
		if errors.Is(err, usecase.ErrURLDuplicate) {
			return nil, status.Error(codes.AlreadyExists, "url already exists")
		} else {
			return nil, status.Error(codes.Unknown, "url saving failed")
		}
	}

	response := &pb.ShortenURLResponse{
		Result: s.urlUC.BuildRedirectURL(url),
	}

	return response, nil
}

// TODO: add url parsing
func (s *ShortenerServer) ResolveURL(ctx context.Context, in *pb.ResolveURLRequest) (*pb.ResolveURLResponse, error) {
	if len(in.Url) == 0 {
		return nil, status.Error(codes.InvalidArgument, "invalid url")
	}

	url, err := s.urlUC.ResolveURL(ctx, in.Url)
	if err != nil || url.Deleted {
		return nil, status.Error(codes.NotFound, "url resolving failed")
	}

	response := &pb.ResolveURLResponse{
		Result: url.Original,
	}

	return response, nil
}
