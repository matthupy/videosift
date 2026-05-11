package ffmpeg

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// ParseMetadataFile reads an FFmpeg metadata=print sidecar file and returns a
// map of frame index (0-based) to presentation timestamp in seconds.
//
// The sidecar format emits one block per frame separated by blank lines:
//
//	frame:0  pts:0  pts_time:0.000
//	lavfi.scene_score=0.512
//
// The frame index matches the %05d suffix in the corresponding PNG filename
// when FFmpeg is invoked with -start_number 0.
func ParseMetadataFile(path string) (map[int]float64, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open metadata file: %w", err)
	}
	defer f.Close()

	result := make(map[int]float64)
	scanner := bufio.NewScanner(f)

	var (
		frameIdx   int
		ptsTime    float64
		hasFrame   bool
		hasPTS     bool
	)

	flush := func() {
		if hasFrame && hasPTS {
			result[frameIdx] = ptsTime
		}
		hasFrame = false
		hasPTS = false
	}

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			flush()
			continue
		}
		// Header line: "frame:0  pts:0  pts_time:0.000  ..."
		if strings.HasPrefix(line, "frame:") {
			flush()
			for _, field := range strings.Fields(line) {
				if strings.HasPrefix(field, "frame:") {
					idx, err := strconv.Atoi(strings.TrimPrefix(field, "frame:"))
					if err == nil {
						frameIdx = idx
						hasFrame = true
					}
				}
				if strings.HasPrefix(field, "pts_time:") {
					t, err := strconv.ParseFloat(strings.TrimPrefix(field, "pts_time:"), 64)
					if err == nil {
						ptsTime = t
						hasPTS = true
					}
				}
			}
		}
	}
	flush()

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan metadata file: %w", err)
	}
	return result, nil
}

// ParseRationalFPS parses an FFmpeg rational frame-rate string such as
// "30000/1001" or "25/1" and returns the result as a float64.
func ParseRationalFPS(s string) (float64, error) {
	parts := strings.SplitN(s, "/", 2)
	if len(parts) != 2 {
		return 0, fmt.Errorf("invalid rational fps %q", s)
	}
	num, err := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
	if err != nil {
		return 0, fmt.Errorf("invalid fps numerator %q: %w", parts[0], err)
	}
	den, err := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
	if err != nil {
		return 0, fmt.Errorf("invalid fps denominator %q: %w", parts[1], err)
	}
	if den == 0 {
		return 0, fmt.Errorf("fps denominator is zero in %q", s)
	}
	return num / den, nil
}
