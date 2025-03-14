package interceptors_test

import (
	"context"
	"log"
	"net"
	"testing"

	"github.com/llravell/go-shortener/internal/entity"
	"github.com/llravell/go-shortener/internal/grpc/interceptors"
	pb "github.com/llravell/go-shortener/internal/proto"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
)

const bufSize = 1024 * 1024

type echoServer struct {
	pb.UnimplementedEchoServer
}

func (s *echoServer) Send(ctx context.Context, in *pb.Message) (*pb.Message, error) {
	return in, nil
}

func startGRPCServer(
	t *testing.T,
	interceptor grpc.UnaryServerInterceptor,
) (pb.EchoClient, func()) {
	t.Helper()

	echo := &echoServer{}

	lis := bufconn.Listen(bufSize)
	s := grpc.NewServer(grpc.UnaryInterceptor(interceptor))
	pb.RegisterEchoServer(s, echo)

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

	return pb.NewEchoClient(conn), closeFn
}

func TestAuth_CheckJWTInterceptor(t *testing.T) {
	log := zerolog.Nop()
	protectedMethods := map[string]bool{
		pb.Echo_Send_FullMethodName: true,
	}
	auth := interceptors.NewAuth("secret", protectedMethods, &log)

	client, closeFn := startGRPCServer(t, auth.CheckJWTInterceptor)
	defer closeFn()

	t.Run("interceptor ignore token checking if method is not protected", func(t *testing.T) {
		protectedMethods[pb.Echo_Send_FullMethodName] = false
		defer func() {
			protectedMethods[pb.Echo_Send_FullMethodName] = true
		}()

		_, err := client.Send(context.Background(), &pb.Message{Text: "test"})

		assert.Nil(t, err)
	})

	t.Run("interceptor return error without token", func(t *testing.T) {
		_, err := client.Send(context.Background(), &pb.Message{Text: "test"})

		assert.Equal(t, codes.Unauthenticated, status.Code(err))
	})

	t.Run("interceptor invoke original method with valid token", func(t *testing.T) {
		jwtToken, err := entity.BuildJWTString("test-uuid", []byte("secret"))
		require.NoError(t, err)

		outgoingMd := metadata.Pairs("token", jwtToken)
		ctxWithToken := metadata.NewOutgoingContext(context.Background(), outgoingMd)

		_, err = client.Send(ctxWithToken, &pb.Message{Text: "test"})

		assert.Nil(t, err)
	})
}
