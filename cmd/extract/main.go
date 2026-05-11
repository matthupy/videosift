package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/matthupy/videosift"
)

func main() {
	var (
		input       = flag.String("i", "", "input video file (required)")
		outputDir   = flag.String("o", "frames", "output directory for extracted PNGs")
		threshold   = flag.Float64("threshold", 0.4, "scene change detection threshold [0.0–1.0]")
		interval    = flag.Float64("interval", 2.0, "temporal sampling interval in seconds")
		maxFrames   = flag.Int("max-frames", 0, "maximum number of output frames (0 = unlimited)")
		hamming     = flag.Int("hamming", 8, "Hamming distance dedup threshold (0 = disabled)")
		algo        = flag.String("algo", "phash", "perceptual hash algorithm: phash or dhash")
		noScene     = flag.Bool("no-scene", false, "disable scene change detection")
		noKeyframe  = flag.Bool("no-keyframe", false, "disable keyframe extraction")
		noTemporal  = flag.Bool("no-temporal", false, "disable temporal sampling")
		noDecimate  = flag.Bool("no-mpdecimate", false, "disable MPDecimate")
		ffmpegPath  = flag.String("ffmpeg", "ffmpeg", "path to ffmpeg binary")
		ffprobePath = flag.String("ffprobe", "ffprobe", "path to ffprobe binary")
		asJSON      = flag.Bool("json", false, "print frame metadata as JSON to stdout")
	)
	flag.Parse()

	if *input == "" {
		fmt.Fprintln(os.Stderr, "error: -i <input> is required")
		flag.Usage()
		os.Exit(1)
	}

	cfg := videosift.DefaultConfig()
	cfg.WorkDir = *outputDir
	cfg.SceneThreshold = *threshold
	cfg.TemporalInterval = *interval
	cfg.MaxFrames = *maxFrames
	cfg.HammingThreshold = *hamming
	cfg.HashAlgo = videosift.HashAlgo(*algo)
	cfg.Scene = !*noScene
	cfg.Keyframe = !*noKeyframe
	cfg.Temporal = !*noTemporal
	cfg.MPDecimate = !*noDecimate
	cfg.FFmpegPath = *ffmpegPath
	cfg.FFprobePath = *ffprobePath

	if err := os.MkdirAll(*outputDir, 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "error: create output dir: %v\n", err)
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	frames, err := videosift.Extract(ctx, *input, cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	if *asJSON {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		if err := enc.Encode(frames); err != nil {
			fmt.Fprintf(os.Stderr, "error: encode JSON: %v\n", err)
			os.Exit(1)
		}
		return
	}

	fmt.Printf("Extracted %d frames to %s\n", len(frames), *outputDir)
	for _, f := range frames {
		fmt.Printf("  [%02d] %6.3fs  %-12s  %s\n", f.Index, f.TimestampSec, f.Strategy, f.Path)
	}
}
