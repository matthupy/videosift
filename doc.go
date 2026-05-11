// Package videosift extracts the minimum set of visually unique frames from a
// video file, suitable for processing with image-based machine learning models
// such as content moderation classifiers.
//
// The package runs up to four FFmpeg-based extraction strategies in parallel
// (scene change detection, keyframe extraction, uniform temporal sampling, and
// MPDecimate), then deduplicates the full candidate set using perceptual
// hashing and Hamming distance comparison.
//
// FFmpeg and FFprobe must be installed and available on PATH (or specified via
// Config.FFmpegPath / Config.FFprobePath). No CGo is required.
//
// Basic usage:
//
//	frames, err := videosift.Extract(ctx, "input.mp4", videosift.DefaultConfig())
//	for _, f := range frames {
//	    fmt.Printf("[%s] %.3fs → %s\n", f.Strategy, f.TimestampSec, f.Path)
//	}
package videosift
