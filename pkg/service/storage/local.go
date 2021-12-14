// Package storage is getting video from different storages
package storage

import (
	"context"
	"fmt"
	"os"

	"go.uber.org/zap"
)

// LocalStorage used for local storage
type LocalStorage struct {
	logger *zap.Logger
}

// Download function returns video from local machine
func (s *LocalStorage) Download(ctx context.Context, id string) (string, error) {
	root := os.Getenv("ROOT")
	videoPath := root + fmt.Sprintf("/tmp/original_video/%s", id)

	if _, err := os.Open(videoPath); err != nil {
		s.logger.Error("Can't find video by id", zap.String("Error", err.Error()))

		return "", err
	}

	return videoPath, nil
}

// NewLocalStorage initialize LocalStorage
func NewLocalStorage(logger *zap.Logger) *LocalStorage {
	return &LocalStorage{
		logger: logger,
	}
}
