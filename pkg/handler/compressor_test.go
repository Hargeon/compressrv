package handler

import (
	"context"
	"github.com/Hargeon/compressrv/pkg/proto"
	"github.com/Hargeon/compressrv/pkg/service"
	"github.com/Hargeon/compressrv/pkg/service/storage"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
	"log"
	"net"
	"os"
	"testing"
)

func init() {
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalln("Can't initialize zap package in compressor_test", err)
	}

	err = godotenv.Load("../../.env")
	if err != nil {
		logger.Fatal("Can't read .env file", zap.String("Error", err.Error()))
	}
	root := os.Getenv("ROOT")
	if root == "" {
		logger.Fatal("Invalid ROOT ENV variable")
	}
}

func dialer() func(context.Context, string) (net.Conn, error) {
	listener := bufconn.Listen(1024 * 1024)
	server := grpc.NewServer()

	logger := zap.NewExample()
	localStorage := storage.NewLocalStorage(logger)
	srv := service.NewService(localStorage)
	handler := NewCompressorHandler(srv)
	proto.RegisterCompressorServer(server, handler)

	go func() {
		if err := server.Serve(listener); err != nil {
			log.Fatalln(err)
		}
	}()

	return func(context.Context, string) (net.Conn, error) {
		return listener.Dial()
	}
}

func TestCompress(t *testing.T) {
	cases := []struct {
		name string
		req  *proto.CompressRequest
	}{
		{
			name: "Check",
			req: &proto.CompressRequest{
				Bitrate:        64000,
				Resolution:     "800:600",
				Ratio:          "4:3",
				VideoServiceId: "test_video.mkv",
			},
		},
	}

	ctx := context.Background()

	conn, err := grpc.DialContext(ctx, "", grpc.WithInsecure(), grpc.WithContextDialer(dialer()))
	if err != nil {
		log.Fatalln(err)
	}
	defer conn.Close()
	client := proto.NewCompressorClient(conn)

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			client.Compress(ctx, testCase.req)
			//if err
		})
	}
}
