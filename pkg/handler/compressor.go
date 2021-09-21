package handler

import (
	"context"
	"github.com/Hargeon/compressrv/pkg/proto"
)

type CompressorHandler struct{}

func NewCompressorHandler() *CompressorHandler {
	return &CompressorHandler{}
}

func (s *CompressorHandler) Compress(ctx context.Context, req *proto.CompressRequest) (*proto.CompressResponse, error) {

	return &proto.CompressResponse{Code: 1}, nil
}
