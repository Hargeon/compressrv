package compressor

import (
	"errors"
	"fmt"
	"github.com/Hargeon/compressrv/pkg/proto"
	"github.com/floostack/transcoder/ffmpeg"
	"os"
	"strconv"
	"strings"
)

const maxRetryCount = 20

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
	newVideoPath := fmt.Sprintf("%s/tmp/converted_video%s", root, newVideoName)
	err := convertVideo(originalVideo, newVideoPath, ffmpegCnf, opts)
	if err != nil {
		return "", err
	}
	return newVideoPath, nil
}

// convertWithBitrate uses for changing bitrate on video file.
func convertWithBitrate(originalVideo string, opts *ffmpeg.Options, ffmpegCnf *ffmpeg.Config) (string, error) {
	successStep := -1
	bitrateVersions := make(map[int]int, maxRetryCount)
	expectedBitrate := *opts.BufferSize

	root := os.Getenv("ROOT")
	iName := strings.LastIndex(originalVideo, "/")

	newVideoName := originalVideo[iName+1:]
	var newVideoPath string

	// Bitrate and buffer size needs for changing bitrate on video file.
	// Buffer size changes on each step and creates new video file.
	for i := 1; i <= maxRetryCount; i++ {
		newVideoPath = fmt.Sprintf("%s/tmp/converted_video/v%d_%s", root, i, newVideoName)

		err := convertVideo(originalVideo, newVideoPath, ffmpegCnf, opts)
		if err != nil {
			continue
		}

		bitrate, err := videoBitrate(newVideoPath, ffmpegCnf)
		if err != nil {
			continue
		}

		bitrateVersions[i] = bitrate
		if bitrate+1000 < expectedBitrate {
			newRate := *opts.BufferSize * (i + 1)
			opts.BufferSize = &newRate
		} else if bitrate-1000 > expectedBitrate {
			newRate := int(float32(*opts.BufferSize) * 0.8)
			opts.BufferSize = &newRate
		} else {
			successStep = i
			break
		}
	}
	if successStep == -1 {
		bIndex, err := findBestBitrate(expectedBitrate, bitrateVersions)
		removeUselessVideos(root, newVideoName, maxRetryCount, bIndex)
		if err == nil {
			return fmt.Sprintf("%s/tmp/converted_video/v%d_%s", root, bIndex, newVideoName), nil
		}
		return "", errors.New("can't convert video file")
	}
	removeUselessVideos(root, newVideoName, successStep, successStep)
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
		bStr := bitrateStr(opt.Bitrate, 0)
		opts.BufferSize = &bufSize
		opts.VideoBitRate = &bStr
	}
	return &opts
}

// bitrateStr uses to get string representation for bitrate.
// Example: input 640000, output: 64k
func bitrateStr(bitrate int32, step int) string {
	if bitrate%1000 == 0 {
		nBitrate := bitrate / 1000
		step++
		return bitrateStr(nBitrate, step)
	} else {
		var bitrateLetter string
		switch step {
		case 1:
			bitrateLetter = "k"
		case 2:
			bitrateLetter = "m"
		case 3:
			bitrateLetter = "g"
		}
		return fmt.Sprintf("%d%s", bitrate, bitrateLetter)
	}
}

func videoBitrate(videoPath string, ffmpegCnf *ffmpeg.Config) (int, error) {
	metaData, err := ffmpeg.New(ffmpegCnf).Input(videoPath).GetMetadata()
	if err != nil {
		return 0, err
	}

	bStr := metaData.GetFormat().GetBitRate()
	return strconv.Atoi(bStr)
}

// removeUselessVideos removes video files which has 'bad' bitrate
func removeUselessVideos(root, videoName string, lastIndex, escapeIndex int) {
	for i := 1; i <= lastIndex; i++ {
		if i == escapeIndex {
			continue
		}
		path := fmt.Sprintf("%s/tmp/converted_video/v%d_%s", root, i, videoName)

		_, err := os.Stat(path)
		if err != nil {
			continue
		}

		os.Remove(path)
	}
}

func findBestBitrate(expectedBitrate int, bitrateVersions map[int]int) (int, error) {
	bestIndex := -1
	bestDifference := -1
	var difference int
	for k, v := range bitrateVersions {
		difference = expectedBitrate - v
		if difference < 0 {
			difference *= -1
		}
		if difference == -1 || difference < bestDifference {
			bestIndex = k
			bestDifference = difference
		}
	}
	if bestIndex == -1 {
		return bestIndex, errors.New("something went wrong while converting video")
	}
	return bestIndex, nil
}