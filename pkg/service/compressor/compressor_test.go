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

func TestFindBestBitrate(t *testing.T) {
	cases := []struct {
		name            string
		expectedBitrate int
		bitrateVersions map[int]int
		expectedIndex   int
		errorPresent    bool
	}{
		{
			name:            "bitrateVersions is empty",
			expectedBitrate: 64000,
			bitrateVersions: map[int]int{},
			expectedIndex:   -1,
			errorPresent:    true,
		},
		{
			name:            "one bitrateVersions pair",
			expectedBitrate: 64000,
			bitrateVersions: map[int]int{4: 35000},
			expectedIndex:   4,
			errorPresent:    false,
		},
		{
			name:            "bitrateVersions less then expected bitrate",
			expectedBitrate: 64000,
			bitrateVersions: map[int]int{1: 10000, 2: 63000, 3: 66000},
			expectedIndex:   2,
			errorPresent:    false,
		},
		{
			name:            "bitrateVersions biggest then expected bitrate",
			expectedBitrate: 64000,
			bitrateVersions: map[int]int{1: 65000, 2: 10000, 3: 38000, 4: 62000},
			expectedIndex:   1,
			errorPresent:    false,
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			index, err := findBestBitrate(testCase.expectedBitrate, testCase.bitrateVersions)
			if index != testCase.expectedIndex {
				t.Errorf("Invalid Index, expected: %d, got: %d\n", testCase.expectedIndex, index)
			}
			if err != nil && !testCase.errorPresent {
				t.Errorf("Unexpected error %v\n", err)
			}
			if err == nil && testCase.errorPresent {
				t.Errorf("Should be error")
			}
		})
	}
}

func TestRemoveUselessVideos(t *testing.T) {
	cases := []struct {
		name        string
		videoName   string
		lastIndex   int
		escapeIndex int
		removeCount int

		// creates maxRetryCount files if set true
		createAllFiles bool
	}{
		{
			name:        "With lastIndex==1, without escapeIndex",
			videoName:   "test_video.mkv",
			lastIndex:   1,
			escapeIndex: -1,
			removeCount: 1,
		},
		{
			name:        "With lastIndex==20, without escapeIndex",
			videoName:   "test_video.mkv",
			lastIndex:   20,
			escapeIndex: -1,
			removeCount: 20,
		},
		{
			name:        "With lastIndex==-1, without escapeIndex",
			videoName:   "test_video.mkv",
			lastIndex:   -1,
			escapeIndex: -1,
			removeCount: 20,
		},
		{
			name:        "With lastIndex==-1, with escapeIndex==5",
			videoName:   "test_video.mkv",
			lastIndex:   -1,
			escapeIndex: 5,
			removeCount: 19,
		},
		{
			name:        "With lastIndex==6, with escapeIndex==2",
			videoName:   "test_video.mkv",
			lastIndex:   6,
			escapeIndex: 2,
			removeCount: 5,
		},
	}

	root := os.Getenv("ROOT")
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			var lastIndex int
			if testCase.lastIndex == -1 {
				lastIndex = maxRetryCount
			} else {
				lastIndex = testCase.lastIndex
			}
			filesPath := createFiles(t, root, testCase.videoName, lastIndex)
			removeUselessVideos(root, testCase.videoName, testCase.lastIndex, testCase.escapeIndex)

			var filesDeleted int
			for _, path := range filesPath {
				if _, err := os.Stat(path); os.IsNotExist(err) {
					filesDeleted++
				}
			}
			if testCase.removeCount != filesDeleted {
				t.Errorf("Should remove %d files, removed %d files\n", testCase.removeCount, filesDeleted)
			}
		})
	}
	clearConvertedVideosDir()
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
			path := fmt.Sprintf("%s/tmp/original_video/%s", root, testCase.videoName)
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
				t.Errorf("Should be error")
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

func createFiles(t *testing.T, root, videoName string, lastIndex int) []string {
	filesPath := make([]string, 0, lastIndex)
	for i := 1; i <= lastIndex; i++ {
		path := fmt.Sprintf("%s%s/v%d_%s", root, convertedVideosPath, i, videoName)
		_, err := os.Create(path)
		if err != nil {
			t.Errorf("Can't create file for test, error: %s\n", err)
			continue
		}
		filesPath = append(filesPath, path)
	}
	return filesPath
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
