package main

import (
	"github.com/Hargeon/compressrv/pkg/handler"
	"github.com/Hargeon/compressrv/pkg/proto"
	"google.golang.org/grpc"
	"log"
	"net"
)

const port = ":8000"

func main() {
	s := grpc.NewServer()
	srv := handler.NewCompressorHandler()
	proto.RegisterCompressorServer(s, srv)

	l, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalln(err)
	}

	log.Println("Starting grpc server...")
	if err := s.Serve(l); err != nil {
		log.Fatalln(err)
	}
}
