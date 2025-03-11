package server_test

import (
	"context"
	"log"
	"net"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/llravell/go-shortener/internal/entity"
	"github.com/llravell/go-shortener/internal/grpc/server"
	"github.com/llravell/go-shortener/internal/mocks"
	pb "github.com/llravell/go-shortener/internal/proto"
	"github.com/llravell/go-shortener/internal/usecase"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

const bufSize = 1024 * 1024

var lis *bufconn.Listener

func bufDialer(context.Context, string) (net.Conn, error) {
	return lis.Dial()
}

func shortenerClient(t *testing.T) (pb.ShortenerClient, *grpc.ClientConn) {
	conn, err := grpc.DialContext(
		context.Background(),
		"bufnet",
		grpc.WithContextDialer(bufDialer),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err)

	return pb.NewShortenerClient(conn), conn
}

func TestShortenerServer(t *testing.T) {
	gen := mocks.NewMockHashGenerator(gomock.NewController(t))
	repo := mocks.NewMockURLRepo(gomock.NewController(t))
	wp := mocks.NewMockURLDeleteWorkerPool(gomock.NewController(t))
	logger := zerolog.Nop()

	gen.EXPECT().Generate().AnyTimes()
	urlUseCase := usecase.NewURLUseCase(repo, wp, gen, "http://localhost:8080", logger)
	shortenerServer := server.NewShortenerServer(urlUseCase, &logger)

	lis = bufconn.Listen(bufSize)
	s := grpc.NewServer()
	pb.RegisterShortenerServer(s, shortenerServer)

	go func() {
		if err := s.Serve(lis); err != nil {
			log.Fatalf("Server exited with error: %v", err)
		}
	}()

	client, conn := shortenerClient(t)
	defer conn.Close()

	t.Run("shorten url", func(t *testing.T) {
		repo.EXPECT().
			Store(gomock.Any(), gomock.Any()).
			Return(&entity.URL{Short: "a"}, nil)

		resp, err := client.ShortenURL(context.Background(), &pb.ShortenURLRequest{
			Url: "https://a.ru",
		})

		require.NoError(t, err)
		assert.Equal(t, "http://localhost:8080/a", resp.Result)
	})

	t.Run("resolve url", func(t *testing.T) {
		repo.EXPECT().
			GetURL(gomock.Any(), "a").
			Return(&entity.URL{Original: "https://a.ru"}, nil)

		resp, err := client.ResolveURL(context.Background(), &pb.ResolveURLRequest{
			Url: "a",
		})

		require.NoError(t, err)
		assert.Equal(t, "https://a.ru", resp.Result)
	})
}
