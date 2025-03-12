package server_test

import (
	"context"
	"errors"
	"log"
	"net"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/llravell/go-shortener/internal/entity"
	"github.com/llravell/go-shortener/internal/grpc/server"
	"github.com/llravell/go-shortener/internal/mocks"
	pb "github.com/llravell/go-shortener/internal/proto"
	repository "github.com/llravell/go-shortener/internal/repo"
	"github.com/llravell/go-shortener/internal/usecase"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
)

const bufSize = 1024 * 1024

var errSomethingHappened = errors.New("something happened")

func startGRPCServer(t *testing.T, repo *mocks.MockURLRepo) (pb.ShortenerClient, func()) {
	t.Helper()

	gen := mocks.NewMockHashGenerator(gomock.NewController(t))
	wp := mocks.NewMockURLDeleteWorkerPool(gomock.NewController(t))
	logger := zerolog.Nop()

	gen.EXPECT().Generate().AnyTimes()
	urlUseCase := usecase.NewURLUseCase(repo, wp, gen, "http://localhost:8080", logger)
	shortenerServer := server.NewShortenerServer(urlUseCase, &logger)

	lis := bufconn.Listen(bufSize)
	s := grpc.NewServer()
	pb.RegisterShortenerServer(s, shortenerServer)

	go func() {
		if err := s.Serve(lis); err != nil {
			log.Fatalf("Server exited with error: %v", err)
		}
	}()

	conn, err := grpc.DialContext(
		context.Background(),
		"bufnet",
		grpc.WithContextDialer(func(ctx context.Context, s string) (net.Conn, error) {
			return lis.Dial()
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err)

	closeFn := func() {
		conn.Close()
		s.Stop()
	}

	return pb.NewShortenerClient(conn), closeFn
}

func TestShortenerServer_ShortenURL(t *testing.T) {
	repo := mocks.NewMockURLRepo(gomock.NewController(t))
	client, closeFn := startGRPCServer(t, repo)

	defer closeFn()

	testCases := map[string]struct {
		url       string
		result    string
		mock      func()
		errorCode codes.Code
	}{
		"return short url": {
			url:    "https://a.ru",
			result: "http://localhost:8080/a",
			mock: func() {
				repo.EXPECT().
					Store(gomock.Any(), gomock.Any()).
					Return(&entity.URL{Short: "a"}, nil)
			},
		},
		"already exists error": {
			url:       "https://a.ru",
			errorCode: codes.AlreadyExists,
			mock: func() {
				repo.EXPECT().
					Store(gomock.Any(), gomock.Any()).
					Return(nil, repository.ErrOriginalURLConflict)
			},
		},
		"url saving error": {
			url:       "https://a.ru",
			errorCode: codes.Unknown,
			mock: func() {
				repo.EXPECT().
					Store(gomock.Any(), gomock.Any()).
					Return(nil, errSomethingHappened)
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			tc.mock()

			resp, err := client.ShortenURL(context.Background(), &pb.ShortenURLRequest{
				Url: tc.url,
			})

			if tc.errorCode > 0 {
				assert.Equal(t, tc.errorCode, status.Code(err))
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.result, resp.Result)
			}
		})
	}
}

func TestShortenerServer_ResolveURL(t *testing.T) {
	repo := mocks.NewMockURLRepo(gomock.NewController(t))
	client, closeFn := startGRPCServer(t, repo)

	defer closeFn()

	testCases := map[string]struct {
		url       string
		result    string
		mock      func()
		errorCode codes.Code
	}{
		"resolve url": {
			url:    "http://localhost:8080/a",
			result: "https://a.ru",
			mock: func() {
				repo.EXPECT().
					GetURL(gomock.Any(), "a").
					Return(&entity.URL{Original: "https://a.ru"}, nil)
			},
		},
		"url resolving error": {
			url:       "http://localhost:8080/a",
			errorCode: codes.NotFound,
			mock: func() {
				repo.EXPECT().
					GetURL(gomock.Any(), "a").
					Return(nil, errSomethingHappened)
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			tc.mock()

			resp, err := client.ResolveURL(context.Background(), &pb.ResolveURLRequest{
				Url: tc.url,
			})

			if tc.errorCode > 0 {
				assert.Equal(t, tc.errorCode, status.Code(err))
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.result, resp.Result)
			}
		})
	}
}
