package handler

import (
	"context"
	"fmt"
	"github.com/Hargeon/compressrv/pkg/proto"
	"github.com/Hargeon/compressrv/pkg/service"
)

type CompressorHandler struct {
	service *service.Service
}

func NewCompressorHandler(s *service.Service) *CompressorHandler {
	return &CompressorHandler{service: s}
}

func (s *CompressorHandler) Compress(ctx context.Context, req *proto.CompressRequest) (*proto.CompressResponse, error) {
	videoPath, err := s.service.VideoStorage.VideoById(req.VideoServiceId)
	if err != nil {
		return &proto.CompressResponse{Code: 0}, err
	}
	fmt.Println(videoPath)
	nVideo, err := s.service.Compressor.Convert(req, videoPath)
	if err != nil {
		return &proto.CompressResponse{Code: 0}, err
	}
	fmt.Println(nVideo)
	return &proto.CompressResponse{Code: 1}, nil
}
