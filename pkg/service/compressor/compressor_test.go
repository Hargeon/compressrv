package compressor

import "testing"

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
