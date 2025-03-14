package interceptors_test

import (
	"context"
	"log"
	"net"
	"testing"

	pb "github.com/llravell/go-shortener/internal/proto"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
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
