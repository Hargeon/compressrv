package handler

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"reflect"
	"testing"

	"github.com/Hargeon/compressrv/pkg/response"
	"github.com/Hargeon/compressrv/pkg/service"
	"github.com/Hargeon/compressrv/pkg/service/compressor"

	"github.com/joho/godotenv"
	"go.uber.org/zap"
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

	if root := os.Getenv("ROOT"); root == "" {
		logger.Fatal("Invalid ROOT ENV variable")
	}
}

type errorCloud struct{}

func (s *errorCloud) Download(ctx context.Context, id string) (string, error) {
	return "", errors.New("mock failed")
}

func (s *errorCloud) Upload(ctx context.Context, fileName string, file io.Reader) (string, error) {
	return "", errors.New("mock failed")
}

type successCloud struct{}

func (s *successCloud) Download(ctx context.Context, id string) (string, error) {
	src := fmt.Sprintf("%s/tmp/original_video/test_video.mkv", os.Getenv("ROOT"))
	dst := fmt.Sprintf("%s/tmp/converted_video/temp_original_file.mkv", os.Getenv("ROOT"))
	sourceFileStat, err := os.Stat(src)

	if err != nil {
		return "", err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return "", fmt.Errorf("%s is not a regular file", src)
	}

	source, err := os.Open(src)
	if err != nil {
		return "", err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return "", err
	}
	defer destination.Close()

	_, err = io.Copy(destination, source)

	if err != nil {
		return "", err
	}

	return dst, nil
}

func (s *successCloud) Upload(ctx context.Context, fileName string, file io.Reader) (string, error) {
	return "temp_converted_file.mkv", nil
}

type errorCompressService struct{}

func (e *errorCompressService) Convert(ctx context.Context, opt *compressor.Request, originalVideo string) (string, error) {
	return "", errors.New("failed mock file convert")
}

func (e *errorCompressService) VideoInfo(path string) (*response.Video, error) {
	return nil, errors.New("failed mock file info")
}

type successCompressService struct{}

func (s *successCompressService) Convert(ctx context.Context, opt *compressor.Request, originalVideo string) (string, error) {
	src := fmt.Sprintf("%s/tmp/original_video/bitrate.mkv", os.Getenv("ROOT"))
	dst := fmt.Sprintf("%s/tmp/converted_video/temp_converted_file.mkv", os.Getenv("ROOT"))
	sourceFileStat, err := os.Stat(src)

	if err != nil {
		return "", err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return "", fmt.Errorf("%s is not a regular file", src)
	}

	source, err := os.Open(src)
	if err != nil {
		return "", err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return "", err
	}
	defer destination.Close()

	_, err = io.Copy(destination, source)
	if err != nil {
		return "", err
	}

	return dst, nil
}

func (s *successCompressService) VideoInfo(path string) (*response.Video, error) {
	resp := &response.Video{
		Bitrate:     64000,
		ResolutionX: 800,
		ResolutionY: 600,
		RatioX:      4,
		RatioY:      3,
	}

	return resp, nil
}

func TestCompress(t *testing.T) {
	logger := zap.NewExample()

	cases := []struct {
		name             string
		srv              *service.Service
		req              *compressor.Request
		expectedResponse *response.Response
	}{
		{
			name: "Invalid downloading video",
			srv: &service.Service{
				VideoStorage: &errorCloud{},
			},
			req: &compressor.Request{
				UserID:         1,
				RequestID:      1,
				Bitrate:        64000,
				Resolution:     "800:600",
				Ratio:          "4:3",
				VideoID:        1,
				VideoServiceID: "mock_service",
			},
			expectedResponse: &response.Response{
				RequestID: 1,
				Error:     "Can't download original video from cloud",
			},
		},
		{
			name: "Invalid converting video",
			srv: &service.Service{
				VideoStorage: &successCloud{},
				Compressor:   &errorCompressService{},
			},
			req: &compressor.Request{
				UserID:         1,
				RequestID:      1,
				Bitrate:        64000,
				Resolution:     "800:600",
				Ratio:          "4:3",
				VideoID:        1,
				VideoServiceID: "mock_service",
			},
			expectedResponse: &response.Response{
				RequestID: 1,
				Error:     "Error occurred when converting video",
			},
		},
		{
			name: "Valid converting video",
			srv: &service.Service{
				VideoStorage: &successCloud{},
				Compressor:   &successCompressService{},
			},
			req: &compressor.Request{
				UserID:         1,
				RequestID:      1,
				Bitrate:        64000,
				Resolution:     "800:600",
				Ratio:          "4:3",
				VideoID:        1,
				VideoServiceID: "mock_service",
			},
			expectedResponse: &response.Response{
				RequestID: 1,
				OriginalVideo: &response.OriginalVideo{
					ID: 1,
					Video: response.Video{
						Bitrate:     64000,
						ResolutionX: 800,
						ResolutionY: 600,
						RatioX:      4,
						RatioY:      3,
					},
				},
				ConvertedVideo: &response.ConvertedVideo{
					ServiceID: "temp_converted_file.mkv",
					Size:      3595197,
					Name:      "temp_converted_file.mkv",
					UserID:    1,
					Video: response.Video{
						Bitrate:     64000,
						ResolutionX: 800,
						ResolutionY: 600,
						RatioX:      4,
						RatioY:      3,
					},
				},
			},
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			srv := NewHandler(testCase.srv, logger)

			resp := srv.Compress(context.Background(), testCase.req)

			if !reflect.DeepEqual(resp, testCase.expectedResponse) {
				t.Errorf("Expected original video: %v, got: %v\n",
					testCase.expectedResponse.OriginalVideo, resp.OriginalVideo)

				t.Errorf("Expected converted video: %v, got: %v\n",
					testCase.expectedResponse.ConvertedVideo, resp.ConvertedVideo)

				if testCase.expectedResponse.RequestID != resp.RequestID {
					t.Errorf("Expected request id: %d, got: %d\n",
						testCase.expectedResponse.RequestID, resp.RequestID)
				}

				if testCase.expectedResponse.Error != resp.Error {
					t.Errorf("Expected error: %s, got: %s\n",
						testCase.expectedResponse.Error, resp.Error)
				}
			}
		})
	}
}
