package compressor

// Request from rabbit mq
type Request struct {
	RequestID      int64  `json:"request_id"`
	Bitrate        int64  `json:"bitrate"`
	Resolution     string `json:"resolution"`
	Ratio          string `json:"ratio"`
	VideoID        int64  `json:"video_id"`
	VideoServiceID string `json:"video_service_id"`
}
