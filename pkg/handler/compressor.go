// Package handler uses for downloading original video, converting file and uploading converted video
package handler

import (
	"context"
	"os"

	"github.com/Hargeon/compressrv/pkg/response"
	"github.com/Hargeon/compressrv/pkg/service"
	"github.com/Hargeon/compressrv/pkg/service/compressor"

	"go.uber.org/zap"
)

// CompressorHandler uses for compressing video file
type CompressorHandler struct {
	srv    *service.Service
	logger *zap.Logger
}

// NewHandler initialize CompressorHandler
func NewHandler(s *service.Service, logger *zap.Logger) *CompressorHandler {
	return &CompressorHandler{srv: s, logger: logger}
}

// Compress video file and build response
func (h *CompressorHandler) Compress(ctx context.Context, req *compressor.Request) *response.Response {
	resp := &response.Response{RequestID: req.RequestID}
	videoName, err := h.srv.Download(ctx, req.VideoServiceID)

	if err != nil {
		h.logger.Error("Download original video",
			zap.String("Error", err.Error()),
			zap.Int64("VideoID", req.VideoID))

		resp.Error = "Can't download original video from cloud"

		return resp
	}

	defer func() {
		os.Remove(videoName)
	}()

	fileInfo, err := h.srv.VideoInfo(videoName)
	if err == nil {
		resp.OriginalVideo = &response.OriginalVideo{
			ID:    req.VideoID,
			Video: *fileInfo,
		}
	} else {
		h.logger.Error("original video video info",
			zap.String("Error", err.Error()),
			zap.Int64("VideoID", req.VideoID))
	}

	convertedVideoPath, err := h.srv.Convert(ctx, req, videoName)
	if err != nil {
		h.logger.Error("Convert original video",
			zap.String("Error", err.Error()),
			zap.Int64("VideoID", req.VideoID))

		resp.Error = "Error occurred when converting video"

		return resp
	}

	defer func() {
		os.Remove(convertedVideoPath)
	}()

	convertedVideo, err := os.Open(convertedVideoPath)
	if err != nil {
		h.logger.Error("open converted video",
			zap.String("Error", err.Error()),
			zap.Int64("VideoID", req.VideoID))

		resp.Error = "error occurred when reading converted video"

		return resp
	}

	id, err := h.srv.Upload(ctx, req.VideoServiceID, convertedVideo)
	if err != nil {
		h.logger.Error("upload converted video",
			zap.String("Error", err.Error()),
			zap.Int64("VideoID", req.VideoID))

		resp.Error = "error occurred when uploading converted video"

		return resp
	}

	fileInfo, err = h.srv.VideoInfo(convertedVideoPath)
	if err != nil {
		h.logger.Error("converted video video info",
			zap.String("Error", err.Error()),
			zap.Int64("VideoID", req.VideoID))

		resp.Error = "error occurred when getting stats converted video"

		return resp
	}

	resp.ConvertedVideo = &response.ConvertedVideo{
		ServiceID: id,
		Video:     *fileInfo,
		Name:      id,
		UserID:    req.UserID,
	}

	stat, err := convertedVideo.Stat()
	if err == nil {
		resp.ConvertedVideo.Size = stat.Size()
	} else {
		h.logger.Error("error occurred when getting size of converted video",
			zap.String("Error", err.Error()),
			zap.Int64("VideoID", req.VideoID))
	}

	return resp
}
