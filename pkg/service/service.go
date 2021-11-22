package service

import (
	"context"
	"io"

	"github.com/Hargeon/compressrv/pkg/response"
	"github.com/Hargeon/compressrv/pkg/service/compressor"
)

type VideoStorage interface {
	Download(ctx context.Context, id string) (string, error)
	Upload(ctx context.Context, fileName string, file io.Reader) (string, error)
}

type Compressor interface {
	Convert(ctx context.Context, opt *compressor.Request, originalVideo string) (string, error)
	VideoInfo(path string) (*response.Video, error)
}

type Service struct {
	VideoStorage
	Compressor
}

func NewService(storage VideoStorage, ffmpegPath, ffprobePath string) *Service {
	return &Service{
		VideoStorage: storage,
		Compressor:   compressor.NewCompressor(ffmpegPath, ffprobePath),
	}
}
