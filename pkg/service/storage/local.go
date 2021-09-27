// Package storage is getting video from different storages
package storage

import (
	"fmt"
	"go.uber.org/zap"
	"os"
)

// LocalStorage used for local storage
type LocalStorage struct {
	logger *zap.Logger
}

// VideoById returns video from local machine
func (s *LocalStorage) VideoById(id string) (string, error) {
	fmt.Println("Get video from storage")
	root := os.Getenv("ROOT")
	videoPath := root + fmt.Sprintf("/tmp/original_video/%s", id)
	_, err := os.Open(videoPath)
	if err != nil {
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
