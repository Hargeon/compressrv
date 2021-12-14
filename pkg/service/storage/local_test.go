package storage

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

func init() {
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalln("Can't initialize zap package in local_test", err)
	}

	err = godotenv.Load("../../../.env")
	if err != nil {
		logger.Fatal("Can't read .env file", zap.String("Error", err.Error()))
	}

	if root := os.Getenv("ROOT"); root == "" {
		logger.Fatal("Invalid ROOT ENV variable")
	}
}

func TestDownload(t *testing.T) {
	root := os.Getenv("ROOT")
	cases := []struct {
		name       string
		input      string
		path       string
		errorExist bool
	}{
		{
			name:       "file is exist",
			input:      "test_video.mkv",
			path:       fmt.Sprintf("%s/tmp/original_video/test_video.mkv", root),
			errorExist: false,
		},
		{
			name:       "File doesn't exist",
			input:      "test_video1.mkv",
			path:       "",
			errorExist: true,
		},
	}

	logger := zap.NewExample()
	storage := NewLocalStorage(logger)
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			path, err := storage.Download(context.Background(), testCase.input)
			if err != nil && !testCase.errorExist {
				t.Errorf("Unexpected error: %v\n", err)
			}
			if err == nil && testCase.errorExist {
				t.Errorf("Should be error\n")
			}
			if path != testCase.path {
				t.Errorf("Invalid path, expected: %s, got: %s\n", testCase.path, path)
			}
		})
	}
}
