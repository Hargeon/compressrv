package main

import (
	"github.com/Hargeon/compressrv/pkg/handler"
	"github.com/Hargeon/compressrv/pkg/proto"
	"github.com/Hargeon/compressrv/pkg/service"
	"github.com/Hargeon/compressrv/pkg/service/storage"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"log"
	"net"
)

const port = ":8000"

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalln(err)
	}
	defer logger.Sync()

	s := grpc.NewServer()
	st := storage.NewLocalStorage(logger)
	srv := service.NewService(st)
	hnd := handler.NewCompressorHandler(srv)
	proto.RegisterCompressorServer(s, hnd)

	l, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalln(err)
	}

	log.Println("Starting grpc server...")
	if err := s.Serve(l); err != nil {
		log.Fatalln(err)
	}
}
