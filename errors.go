package videosift

import (
	"errors"
	"fmt"
)

// FFmpegError wraps a non-zero ffmpeg or ffprobe exit with its captured stderr.
type FFmpegError struct {
	Args   []string
	Stderr string
	Cause  error
}

func (e *FFmpegError) Error() string {
	return fmt.Sprintf("ffmpeg %v: %v\nstderr: %s", e.Args, e.Cause, e.Stderr)
}

func (e *FFmpegError) Unwrap() error { return e.Cause }

var (
	// ErrNoBinaries is returned when ffmpeg or ffprobe cannot be found.
	ErrNoBinaries = errors.New("videosift: ffmpeg/ffprobe not found in PATH")

	// ErrNoStrategies is returned when all extraction strategies are disabled.
	ErrNoStrategies = errors.New("videosift: all extraction strategies disabled")

	// ErrNoFrames is returned when the pipeline produces zero frames.
	ErrNoFrames = errors.New("videosift: no frames extracted")
)
