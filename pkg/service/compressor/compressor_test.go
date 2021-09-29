package compressor

import (
	"fmt"
	"github.com/Hargeon/compressrv/pkg/proto"
	"github.com/floostack/transcoder/ffmpeg"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
	"os"
	"testing"
)

const originalVideoPath = "/tmp/original_video/"

func init() {
	logger := zap.NewExample()

	err := godotenv.Load("../../../.env")
	if err != nil {
		logger.Fatal("Can't read .env file", zap.String("Error", err.Error()))
	}
	root := os.Getenv("ROOT")
	if root == "" {
		logger.Fatal("Invalid ROOT ENV variable")
	}
	ffmpegPath := os.Getenv("FFMPEG_PATH")
	if ffmpegPath == "" {
		logger.Fatal("Invalid FFMPEG_PATH ENV variable")
	}
	ffprobePath := os.Getenv("FFPROBE_PATH")
	if ffprobePath == "" {
		logger.Fatal("Invalid FFMPEG_PATH ENV variable")
	}
}

func TestVideoBitrate(t *testing.T) {
	cases := []struct {
		name            string
		videoName       string
		expectedBitrate int
		errorPresent    bool
	}{
		{
			name:            "test_video.mkv file present",
			videoName:       "test_video.mkv",
			expectedBitrate: 274625,
			errorPresent:    false,
		},
		{
			name:            "bitrate.mkv file present",
			videoName:       "bitrate.mkv",
			expectedBitrate: 484201,
			errorPresent:    false,
		},
		{
			name:            "not_exist.mkv file present",
			videoName:       "not_exist.mkv",
			expectedBitrate: 0,
			errorPresent:    true,
		},
	}

	root := os.Getenv("ROOT")
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			path := fmt.Sprintf("%s%s%s", root, originalVideoPath, testCase.videoName)
			cnf := ffmpeg.Config{
				FfmpegBinPath:  os.Getenv("FFMPEG_PATH"),
				FfprobeBinPath: os.Getenv("FFPROBE_PATH"),
			}

			bitrate, err := videoBitrate(path, &cnf)
			if bitrate != testCase.expectedBitrate {
				t.Errorf("Invalid bitrate, expected: %d, got: %d\n", testCase.expectedBitrate, bitrate)
			}

			if err != nil && !testCase.errorPresent {
				t.Errorf("Unexpected error: %s\n", err)
			}

			if err == nil && testCase.errorPresent {
				t.Errorf("Should be error\n")
			}
		})
	}
}

func TestBuildOptions(t *testing.T) {
	cases := []struct {
		name string
		opts *proto.CompressRequest

		resolution   string
		ration       string
		bufferSize   int
		videoBitrate string
	}{
		{
			name:         "With ratio",
			opts:         &proto.CompressRequest{Ratio: "6:4"},
			resolution:   "",
			ration:       "6:4",
			bufferSize:   0,
			videoBitrate: "",
		},
		{
			name:         "With resolution",
			opts:         &proto.CompressRequest{Resolution: "700:600"},
			resolution:   "700:600",
			ration:       "",
			bufferSize:   0,
			videoBitrate: "",
		},
		{
			name:         "With bitrate",
			opts:         &proto.CompressRequest{Bitrate: 64000},
			resolution:   "",
			ration:       "",
			bufferSize:   64000,
			videoBitrate: "64000",
		},
		{
			name: "With resolution, ration and bitrate",
			opts: &proto.CompressRequest{
				Bitrate:    100000,
				Resolution: "400:300",
				Ratio:      "9:4",
			},
			resolution:   "400:300",
			ration:       "9:4",
			bufferSize:   100000,
			videoBitrate: "100000",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			opts := buildOptions(testCase.opts)

			if opts.Resolution == nil {
				if testCase.resolution != "" {
					t.Errorf("Invalid resolution, expected: %s, got: nil\n", testCase.resolution)
				}
			} else {
				if testCase.resolution != *opts.Resolution {
					t.Errorf("Invalid resolution, expected: %s, got: %s\n", testCase.resolution, *opts.Resolution)
				}
			}

			if opts.Aspect == nil {
				if testCase.ration != "" {
					t.Errorf("Invalid ratio, expected: %s, got: nil\n", testCase.ration)
				}
			} else {
				if testCase.ration != *opts.Aspect {
					t.Errorf("Invalid ratio, expected: %s, got: %s\n", testCase.ration, *opts.Aspect)
				}
			}

			if opts.BufferSize == nil {
				if testCase.bufferSize != 0 {
					t.Errorf("Invalid buffer size, expected: %d, got: nil\n", testCase.bufferSize)
				}
			} else {
				if testCase.bufferSize != *opts.BufferSize {
					t.Errorf("Invalid buffer size, expected: %d, got: %d\n", testCase.bufferSize, *opts.BufferSize)
				}
			}

			if opts.VideoBitRate == nil {
				if testCase.videoBitrate != "" {
					t.Errorf("Invalid buffer size, expected: %s, got: nil\n", testCase.videoBitrate)
				}
			} else {
				if testCase.videoBitrate != *opts.VideoBitRate {
					t.Errorf("Invalid video bitrate, expected: %s, got: %s\n", testCase.videoBitrate, *opts.VideoBitRate)
				}
			}
		})
	}
}

func TestConvertVideo(t *testing.T) {
	root := os.Getenv("ROOT")
	originPath := fmt.Sprintf("%s%stest_video.mkv", root, originalVideoPath)
	newPath := fmt.Sprintf("%s%s/v1_test_video.mkv", root, convertedVideosPath)

	ffmpegCnf := &ffmpeg.Config{
		FfmpegBinPath:  os.Getenv("FFMPEG_PATH"),
		FfprobeBinPath: os.Getenv("FFPROBE_PATH"),
	}

	cases := []struct {
		name      string
		ffmpegCnf *ffmpeg.Config
		opts      *ffmpeg.Options

		inputResolution string
		inputRation     string
		inputBitrate    string
		inputBufferSize int

		expectedResolution string
		expectedRation     string

		errorPresent bool
		fileCreated  bool
	}{
		{
			name:               "With full ffmpeg config",
			ffmpegCnf:          ffmpegCnf,
			inputResolution:    "800:600",
			inputRation:        "4:3",
			inputBitrate:       "64000",
			inputBufferSize:    64000,
			expectedResolution: "800:600",
			expectedRation:     "4:3",
			errorPresent:       false,
		},
		{
			name: "Only with ffprobe path for ffmpeg",
			ffmpegCnf: &ffmpeg.Config{
				FfprobeBinPath: os.Getenv("FFPROBE_PATH"),
			},
			inputResolution:    "800:600",
			inputRation:        "4:3",
			inputBitrate:       "64000",
			inputBufferSize:    64000,
			expectedResolution: "",
			expectedRation:     "",
			errorPresent:       true,
		},
		{
			name: "Only with ffmpeg path for ffmpeg",
			ffmpegCnf: &ffmpeg.Config{
				FfmpegBinPath: os.Getenv("FFMPEG_PATH"),
			},
			inputResolution:    "800:600",
			inputRation:        "4:3",
			inputBitrate:       "64000",
			inputBufferSize:    64000,
			expectedResolution: "",
			expectedRation:     "",
			errorPresent:       true,
		},
	}

	for i, testCase := range cases {
		opts := &ffmpeg.Options{
			Resolution: &testCase.inputResolution,
			Aspect:     &testCase.inputRation,
		}
		testCase.opts = opts
		cases[i] = testCase
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			defer clearConvertedVideosDir()
			err := convertVideo(originPath, newPath, testCase.ffmpegCnf, testCase.opts)
			if err == nil && testCase.errorPresent {
				t.Errorf("Should be error\n")
			}
			if err != nil && !testCase.errorPresent {
				t.Errorf("Unexpected error: %s\n", err)
			}

			if testCase.expectedRation != "" || testCase.expectedResolution != "" {
				metaData, err := ffmpeg.New(ffmpegCnf).Input(newPath).GetMetadata()
				if err != nil {
					t.Errorf("Unexpected error while check video metadata, error: %s\n", err)
					return
				}
				streams := metaData.GetStreams()
				resolution := fmt.Sprintf("%d:%d", streams[0].GetWidth(), streams[0].GetHeight())
				if resolution != testCase.expectedResolution {
					t.Errorf("Invalid resolution, expected: %s, got: %s\n", testCase.expectedResolution, resolution)
				}
				ratio := streams[0].GetDisplayAspectRatio()
				if ratio != testCase.expectedRation {
					t.Errorf("Invalid ration, expected: %s, got: %s\n", testCase.expectedRation, ratio)
				}
			}
		})
	}
}

func TestConvertWithBitrate(t *testing.T) {
	ffmpegCnf := &ffmpeg.Config{
		FfmpegBinPath:  os.Getenv("FFMPEG_PATH"),
		FfprobeBinPath: os.Getenv("FFPROBE_PATH"),
	}

	cases := []struct {
		name          string
		originalVideo string
		inputBuffer   int
		opts          *ffmpeg.Options
		ffmpegCnf     *ffmpeg.Config

		expectedBitrate int
		filePresent     bool
		errorPresent    bool
	}{
		{
			name:            "Convert test_video.mkv to 64000 bit/s",
			originalVideo:   "test_video.mkv",
			ffmpegCnf:       ffmpegCnf,
			inputBuffer:     64000,
			expectedBitrate: 64000,
			filePresent:     true,
			errorPresent:    false,
		},
		{
			name:            "Convert test_video.mkv to 78000 bit/s",
			originalVideo:   "test_video.mkv",
			ffmpegCnf:       ffmpegCnf,
			inputBuffer:     78000,
			expectedBitrate: 78000,
			filePresent:     true,
			errorPresent:    false,
		},
		{
			name:            "Convert test_video.mkv to 30000 bit/s",
			originalVideo:   "test_video.mkv",
			ffmpegCnf:       ffmpegCnf,
			inputBuffer:     30000,
			expectedBitrate: 30000,
			filePresent:     true,
			errorPresent:    false,
		},
		{
			name:            "Convert bitrate.mkv to 100000 bit/s",
			originalVideo:   "bitrate.mkv",
			ffmpegCnf:       ffmpegCnf,
			inputBuffer:     100000,
			expectedBitrate: 100000,
			filePresent:     true,
			errorPresent:    false,
		},
		{
			name:            "Convert bitrate.mkv to 45000 bit/s",
			originalVideo:   "bitrate.mkv",
			ffmpegCnf:       ffmpegCnf,
			inputBuffer:     45000,
			expectedBitrate: 45000,
			filePresent:     true,
			errorPresent:    false,
		},
		{
			name:          "Convert bitrate.mkv to 45000 bit/s with invalid ffmpeg path",
			originalVideo: "bitrate.mkv",
			ffmpegCnf: &ffmpeg.Config{
				FfprobeBinPath: os.Getenv("FFPROBE_PATH"),
			},
			inputBuffer:  45000,
			filePresent:  false,
			errorPresent: true,
		},
		{
			name:          "Convert bitrate.mkv to 45000 bit/s with invalid ffrpobe path",
			originalVideo: "bitrate.mkv",
			ffmpegCnf: &ffmpeg.Config{
				FfmpegBinPath: os.Getenv("FFMPEG_PATH"),
			},
			inputBuffer:  45000,
			filePresent:  false,
			errorPresent: true,
		},
	}

	for i, testCase := range cases {
		buff := testCase.inputBuffer
		bRate := fmt.Sprintf("%d", testCase.inputBuffer)
		opts := &ffmpeg.Options{
			BufferSize:   &buff,
			VideoBitRate: &bRate,
		}
		testCase.opts = opts
		cases[i] = testCase
	}

	root := os.Getenv("ROOT")
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			defer clearConvertedVideosDir()
			originVideoPath := fmt.Sprintf("%s%s%s", root, originalVideoPath, testCase.originalVideo)
			path, err := convertWithBitrate(originVideoPath, testCase.opts, testCase.ffmpegCnf)
			if err != nil && !testCase.errorPresent {
				t.Errorf("Unexpected error, error: %s\n", err)
			}
			if err == nil && testCase.errorPresent {
				t.Errorf("Should be error\n")
			}
			if path == "" && testCase.filePresent {
				t.Errorf("File should be created\n")
			}
			if path != "" && !testCase.filePresent {
				t.Errorf("File should not be created\n")
			}

			if path != "" {
				bitrate, err := videoBitrate(path, ffmpegCnf)
				if err != nil {
					t.Errorf("Unexpected error while checking video bitrate, err: %s\n", err)
					return
				}
				if bitrate+bitrateAccuracy < testCase.expectedBitrate || bitrate-bitrateAccuracy > testCase.expectedBitrate {
					t.Errorf("Invalid bitrate, expected %d +- %d, got %d\n", testCase.expectedBitrate, bitrateAccuracy, bitrate)
				}
			}
		})
	}
}

func clearConvertedVideosDir() {
	root := os.Getenv("ROOT")
	logger := zap.NewExample()
	dirPath := fmt.Sprintf("%s%s/", root, convertedVideosPath)
	dir, err := os.Open(dirPath)
	if err != nil {
		logger.Fatal("Can't open dir", zap.String("Error", err.Error()))
	}
	files, _ := dir.ReadDir(0)
	for _, file := range files {
		fileName := file.Name()
		if fileName == ".keep" {
			continue
		}
		path := fmt.Sprintf("%s%s/%s", root, convertedVideosPath, fileName)
		err := os.Remove(path)
		if err != nil {
			logger.Fatal("Can't remove file", zap.String("Error", err.Error()))
		}
	}
}

func TestConvert(t *testing.T) {
	root := os.Getenv("ROOT")
	cases := []struct {
		name          string
		opt           *proto.CompressRequest
		originalVideo string

		errorPresent       bool
		expectedBitrate    int
		expectedRation     string
		expectedResolution string
	}{
		{
			name:               "Change resolution test_video.mkv",
			opt:                &proto.CompressRequest{Resolution: "800:600"},
			originalVideo:      fmt.Sprintf("%s%stest_video.mkv", root, originalVideoPath),
			errorPresent:       false,
			expectedResolution: "800:600",
		},
		{
			name:           "Change ration test_video.mkv",
			opt:            &proto.CompressRequest{Ratio: "4:3"},
			originalVideo:  fmt.Sprintf("%s%stest_video.mkv", root, originalVideoPath),
			errorPresent:   false,
			expectedRation: "4:3",
		},
		{
			name:            "Change bitrate test_video.mkv",
			opt:             &proto.CompressRequest{Bitrate: 50000},
			originalVideo:   fmt.Sprintf("%s%stest_video.mkv", root, originalVideoPath),
			errorPresent:    false,
			expectedBitrate: 50000,
		},
		{
			name: "Change bitrate, resolution and ratio test_video.mkv",
			opt: &proto.CompressRequest{
				Bitrate:    60000,
				Resolution: "600:300",
				Ratio:      "9:6",
			},
			originalVideo:      fmt.Sprintf("%s%stest_video.mkv", root, originalVideoPath),
			errorPresent:       false,
			expectedBitrate:    60000,
			expectedRation:     "3:2",
			expectedResolution: "600:300",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			defer clearConvertedVideosDir()
			service := NewCompressor()
			path, err := service.Convert(testCase.opt, testCase.originalVideo)
			if err != nil && !testCase.errorPresent {
				t.Errorf("Unexpected error, error: %s\n", err)
			}
			if err == nil && testCase.errorPresent {
				t.Errorf("Should be error\n")
			}

			filePresent := testCase.expectedResolution != "" || testCase.expectedRation != "" || testCase.expectedBitrate != 0
			if filePresent && path == "" {
				t.Errorf("File should be created\n")
			}

			if !filePresent && path != "" {
				t.Errorf("File should not be created\n")
			}

			if path != "" && filePresent {
				ffmpegCnf := &ffmpeg.Config{
					FfmpegBinPath:  os.Getenv("FFMPEG_PATH"),
					FfprobeBinPath: os.Getenv("FFPROBE_PATH"),
				}

				if testCase.expectedBitrate != 0 {
					bitrate, err := videoBitrate(path, ffmpegCnf)
					if err != nil {
						t.Errorf("Unexpected error while checking video bitrate, error: %s\n", err)
					}
					if (bitrate < testCase.expectedBitrate-bitrateAccuracy) || (bitrate > testCase.expectedBitrate+bitrateAccuracy) {
						t.Errorf("Invalid bitrate, expected %d +- %d, got %d\n", testCase.expectedBitrate, bitrateAccuracy, bitrate)
					}
				}

				metaData, err := ffmpeg.New(ffmpegCnf).Input(path).GetMetadata()
				if err != nil {
					t.Errorf("Unexpected error while check video metadata, error: %s\n", err)
					return
				}
				streams := metaData.GetStreams()

				if testCase.expectedResolution != "" {
					resolution := fmt.Sprintf("%d:%d", streams[0].GetWidth(), streams[0].GetHeight())
					if resolution != testCase.expectedResolution {
						t.Errorf("Invalid resolution, expected: %s, got: %s\n", testCase.expectedResolution, resolution)
					}
				}

				if testCase.expectedRation != "" {
					ratio := streams[0].GetDisplayAspectRatio()
					if ratio != testCase.expectedRation {
						t.Errorf("Invalid ration, expected: %s, got: %s\n", testCase.expectedRation, ratio)
					}
				}
			}
		})
	}
}
