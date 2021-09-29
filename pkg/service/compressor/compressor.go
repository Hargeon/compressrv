package compressor

import (
	"fmt"
	"github.com/Hargeon/compressrv/pkg/proto"
	"github.com/floostack/transcoder/ffmpeg"
	"os"
	"strconv"
	"strings"
)

const convertedVideosPath = "/tmp/converted_video"
const bitrateAccuracy = 1000
const decreaseBitrate = 0.6
const increaseBitrate = 2

// Compressor ...
type Compressor struct{}

// NewCompressor ...
func NewCompressor() *Compressor {
	return &Compressor{}
}

// Convert video file originalVideo to new with proto.CompressRequest params
func (c *Compressor) Convert(opt *proto.CompressRequest, originalVideo string) (string, error) {
	ffmpegCnf := &ffmpeg.Config{
		FfmpegBinPath:  os.Getenv("FFMPEG_PATH"),
		FfprobeBinPath: os.Getenv("FFPROBE_PATH"),
	}
	opts := buildOptions(opt)

	if opt.Bitrate != 0 {
		return convertWithBitrate(originalVideo, opts, ffmpegCnf)
	}
	root := os.Getenv("ROOT")
	iName := strings.LastIndex(originalVideo, "/")

	newVideoName := originalVideo[iName:]
	newVideoPath := fmt.Sprintf("%s%s%s", root, convertedVideosPath, newVideoName)
	err := convertVideo(originalVideo, newVideoPath, ffmpegCnf, opts)
	if err != nil {
		return "", err
	}
	return newVideoPath, nil
}

// convertWithBitrate uses for changing bitrate on video file.
func convertWithBitrate(originalVideo string, opts *ffmpeg.Options, ffmpegCnf *ffmpeg.Config) (string, error) {
	expectedBitrate := *opts.BufferSize

	root := os.Getenv("ROOT")
	iName := strings.LastIndex(originalVideo, "/")

	newVideoName := originalVideo[iName+1:]
	var newVideoPath string

	// Bitrate and buffer size needs for changing bitrate on video file.
	// Buffer size changes on each step and creates new video file.
	for i := 1; ; i++ {
		previousVideoPath := fmt.Sprintf("%s%s/v%d_%s", root, convertedVideosPath, i-1, newVideoName)
		newVideoPath = fmt.Sprintf("%s%s/v%d_%s", root, convertedVideosPath, i, newVideoName)

		err := convertVideo(originalVideo, newVideoPath, ffmpegCnf, opts)
		if err != nil {
			return "", err
		}

		bitrate, err := videoBitrate(newVideoPath, ffmpegCnf)
		if err != nil {
			os.Remove(newVideoPath)

			if _, bErr := videoBitrate(previousVideoPath, ffmpegCnf); bErr == nil {
				newVideoPath = previousVideoPath
				break
			} else {
				return "", err
			}
		}

		os.Remove(previousVideoPath)

		if bitrate <= expectedBitrate {
			if expectedBitrate-bitrateAccuracy <= bitrate {
				break
			} else {
				newRate := *opts.BufferSize * increaseBitrate
				opts.BufferSize = &newRate
			}
		} else {
			if expectedBitrate+bitrateAccuracy >= bitrate {
				break
			} else {
				newRate := int(float64(*opts.BufferSize) * decreaseBitrate)
				opts.BufferSize = &newRate
			}
		}
	}
	return newVideoPath, nil
}

func convertVideo(originPath, newPath string, ffmpegCnf *ffmpeg.Config, opts *ffmpeg.Options) error {
	// don't use transcoder.Transcoder in params.
	// Duplicate WithOptions raise error: number of options and output files does not match
	_, err := ffmpeg.
		New(ffmpegCnf).
		Input(originPath).
		Output(newPath).
		WithOptions(*opts).
		Start(*opts)
	return err
}

func buildOptions(opt *proto.CompressRequest) *ffmpeg.Options {
	opts := ffmpeg.Options{}
	if opt.Resolution != "" {
		opts.Resolution = &opt.Resolution
	}
	if opt.Ratio != "" {
		opts.Aspect = &opt.Ratio
	}
	if opt.Bitrate != 0 {
		bufSize := int(opt.Bitrate)
		bStr := fmt.Sprintf("%d", opt.Bitrate)
		opts.BufferSize = &bufSize
		opts.VideoBitRate = &bStr
	}
	return &opts
}

func videoBitrate(videoPath string, ffmpegCnf *ffmpeg.Config) (int, error) {
	metaData, err := ffmpeg.New(ffmpegCnf).Input(videoPath).GetMetadata()
	if err != nil {
		return 0, err
	}

	bStr := metaData.GetFormat().GetBitRate()
	return strconv.Atoi(bStr)
}
