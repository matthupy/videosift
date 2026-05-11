package videosift

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	iffmpeg "github.com/matthupy/videosift/internal/ffmpeg"
)

// probe runs ffprobe on videoPath and returns structured metadata.
func probe(ctx context.Context, videoPath, ffprobePath string) (VideoInfo, error) {
	var lines []string
	cmd := iffmpeg.New(ffprobePath,
		"-v", "quiet",
		"-print_format", "json",
		"-show_streams",
		"-show_format",
		videoPath,
	)
	if err := cmd.RunLines(ctx, func(l string) { lines = append(lines, l) }); err != nil {
		return VideoInfo{}, fmt.Errorf("ffprobe: %w", err)
	}

	var result ffprobeOutput
	if err := json.Unmarshal([]byte(strings.Join(lines, "\n")), &result); err != nil {
		return VideoInfo{}, fmt.Errorf("ffprobe JSON parse: %w", err)
	}

	var info VideoInfo

	// Duration from format section (most reliable).
	if d, err := strconv.ParseFloat(result.Format.Duration, 64); err == nil {
		info.Duration = d
	}
	if br, err := strconv.ParseInt(result.Format.BitRate, 10, 64); err == nil {
		info.BitRate = br
	}

	// Find first video stream.
	for _, s := range result.Streams {
		if s.CodecType != "video" {
			continue
		}
		info.Codec = s.CodecName
		info.Width = s.Width
		info.Height = s.Height

		fps, err := iffmpeg.ParseRationalFPS(s.RFrameRate)
		if err != nil {
			// Fall back to avg_frame_rate.
			fps, _ = iffmpeg.ParseRationalFPS(s.AvgFrameRate)
		}
		info.FrameRate = fps
		break
	}

	return info, nil
}

// ffprobeOutput mirrors the JSON structure returned by ffprobe.
type ffprobeOutput struct {
	Streams []ffprobeStream `json:"streams"`
	Format  ffprobeFormat   `json:"format"`
}

type ffprobeStream struct {
	CodecType    string `json:"codec_type"`
	CodecName    string `json:"codec_name"`
	Width        int    `json:"width"`
	Height       int    `json:"height"`
	RFrameRate   string `json:"r_frame_rate"`
	AvgFrameRate string `json:"avg_frame_rate"`
}

type ffprobeFormat struct {
	Duration string `json:"duration"`
	BitRate  string `json:"bit_rate"`
}
