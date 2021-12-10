// Package response uses for embedded original and converted video params
package response

// Video consists meta data for video
type Video struct {
	Bitrate     int64 `json:"bitrate"`
	ResolutionX int   `json:"resolution_x"`
	ResolutionY int   `json:"resolution_y"`
	RatioX      int   `json:"ratio_x"`
	RatioY      int   `json:"ratio_y"`
}

// OriginalVideo consists fields for original video
type OriginalVideo struct {
	ID int64 `json:"id"`
	Video
}

// ConvertedVideo consists fields for converted video
type ConvertedVideo struct {
	ServiceID string `json:"service_id"`
	Size      int64  `json:"size"`
	Name      string `json:"name"`
	UserID    int64  `json:"user_id"`
	Video
}

// Response represent full response after compressing
type Response struct {
	RequestID      int64           `json:"request_id"`
	OriginalVideo  *OriginalVideo  `json:"original_video,omitempty"`
	ConvertedVideo *ConvertedVideo `json:"converted_video,omitempty"`
	Error          string          `json:"error,omitempty"`
}
