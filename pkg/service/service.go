package service

import (
	"github.com/Hargeon/compressrv/pkg/proto"
	"github.com/Hargeon/compressrv/pkg/service/compressor"
)

type VideoStorage interface {
	VideoById(id string) (string, error)
}

type Compressor interface {
	Convert(opt *proto.CompressRequest, video string) (string, error)
}

type Service struct {
	VideoStorage
	Compressor
}

func NewService(storage VideoStorage) *Service {
	return &Service{
		VideoStorage: storage,
		Compressor:   compressor.NewCompressor(),
	}
}
