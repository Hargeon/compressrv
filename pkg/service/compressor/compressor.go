package compressor

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/Hargeon/compressrv/pkg/response"

	"github.com/floostack/transcoder/ffmpeg"
)

const (
	convertedVideosPath = "/tmp/converted_video"
	bitrateAccuracy     = 1000
	decreaseBitrate     = 0.6
	increaseBitrate     = 2
	ratioNumber         = 2
	decimal             = 10
	bitrateBitSize      = 64
)

// Compressor uses for changing bitrate, resolution and ratio for video
type Compressor struct {
	ffmpegCnf *ffmpeg.Config
}

// NewCompressor initialize Compressor
func NewCompressor(ffmpegPath, ffprobePath string) *Compressor {
	return &Compressor{
		ffmpegCnf: &ffmpeg.Config{
			FfmpegBinPath:  ffmpegPath,
			FfprobeBinPath: ffprobePath,
		},
	}
}

// Convert function change bitrate, resolution and ratio for video
func (c *Compressor) Convert(ctx context.Context, opt *Request, originalVideo string) (string, error) {
	opts := c.buildOptions(opt)

	if opt.Bitrate != 0 {
		return c.convertWithBitrate(originalVideo, opts)
	}

	root := os.Getenv("ROOT")
	iName := strings.LastIndex(originalVideo, "/")

	newVideoName := originalVideo[iName:]
	newVideoPath := fmt.Sprintf("%s%s%s", root, convertedVideosPath, newVideoName)
	err := c.convertVideo(originalVideo, newVideoPath, opts)

	if err != nil {
		return "", err
	}

	return newVideoPath, nil
}

// VideoInfo function calculate bitrate, resolution and ratio for video file
func (c *Compressor) VideoInfo(path string) (*response.Video, error) {
	video := new(response.Video)

	bitrate, err := c.videoBitrate(path)

	if err != nil {
		return nil, err
	}

	video.Bitrate = bitrate

	metaData, err := ffmpeg.New(c.ffmpegCnf).Input(path).GetMetadata()
	if err != nil {
		return nil, err
	}

	streams := metaData.GetStreams()
	video.ResolutionX = streams[0].GetWidth()
	video.ResolutionY = streams[0].GetHeight()

	re, err := regexp.Compile(`[0-9]+`)
	if err != nil {
		return nil, err
	}

	ratioStr := streams[0].GetDisplayAspectRatio() // 4:3
	ratios := re.FindAllString(ratioStr, ratioNumber)
	ratioX, err := strconv.Atoi(ratios[0])

	if err != nil {
		return nil, err
	}

	ratioY, err := strconv.Atoi(ratios[1])
	if err != nil {
		return nil, err
	}

	video.RatioX = ratioX
	video.RatioY = ratioY

	return video, nil
}

// convertWithBitrate uses for changing bitrate for video file.
func (c *Compressor) convertWithBitrate(originalVideo string, opts *ffmpeg.Options) (string, error) {
	expectedBitrate := int64(*opts.BufferSize)

	root := os.Getenv("ROOT")
	iName := strings.LastIndex(originalVideo, "/")

	newVideoName := originalVideo[iName+1:]

	var newVideoPath string

	// Bitrate and buffer size needs for changing bitrate on video file.
	// Buffer size changes on each step and creates new video file.
	for i := 1; ; i++ {
		previousVideoPath := fmt.Sprintf("%s%s/v%d_%s", root, convertedVideosPath, i-1, newVideoName)
		newVideoPath = fmt.Sprintf("%s%s/v%d_%s", root, convertedVideosPath, i, newVideoName)

		err := c.convertVideo(originalVideo, newVideoPath, opts)
		if err != nil {
			return "", err
		}

		bitrate, err := c.videoBitrate(newVideoPath)
		if err != nil {
			os.Remove(newVideoPath)

			if _, bErr := c.videoBitrate(previousVideoPath); bErr == nil {
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

// convertVideo from originPath to newPath with *ffmpeg.Options
func (c *Compressor) convertVideo(originPath, newPath string, opts *ffmpeg.Options) error {
	// don't use transcoder.Transcoder in params.
	// Duplicate WithOptions raise error: number of options and output files does not match
	_, err := ffmpeg.
		New(c.ffmpegCnf).
		Input(originPath).
		Output(newPath).
		WithOptions(*opts).
		Start(*opts)

	return err
}

// buildOptions for converting from *Request
func (c *Compressor) buildOptions(opt *Request) *ffmpeg.Options {
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

// videoBitrate return bitrate of video
func (c *Compressor) videoBitrate(videoPath string) (int64, error) {
	metaData, err := ffmpeg.New(c.ffmpegCnf).Input(videoPath).GetMetadata()
	if err != nil {
		return 0, err
	}

	bStr := metaData.GetFormat().GetBitRate()

	return strconv.ParseInt(bStr, decimal, bitrateBitSize)
}
