package videosift

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	iffmpeg "github.com/matthupy/videosift/internal/ffmpeg"
)

// runScene extracts frames where the scene-change score exceeds cfg.SceneThreshold.
func runScene(ctx context.Context, videoPath string, cfg Config) ([]candidateFrame, error) {
	return runStrategy(ctx, videoPath, cfg, StrategyScene, sceneFilter(cfg))
}

// runKeyframe extracts only I-frames (encoded keyframes).
func runKeyframe(ctx context.Context, videoPath string, cfg Config) ([]candidateFrame, error) {
	return runStrategy(ctx, videoPath, cfg, StrategyKeyframe, keyframeFilter(cfg))
}

// runTemporal extracts one frame every cfg.TemporalInterval seconds.
func runTemporal(ctx context.Context, videoPath string, cfg Config) ([]candidateFrame, error) {
	return runStrategy(ctx, videoPath, cfg, StrategyTemporal, temporalFilter(cfg))
}

// runMPDecimate uses FFmpeg's mpdecimate filter to drop near-duplicate frames.
func runMPDecimate(ctx context.Context, videoPath string, cfg Config) ([]candidateFrame, error) {
	return runStrategy(ctx, videoPath, cfg, StrategyMPDecimate, mpdecimateFilter(cfg))
}

// runStrategy is the shared implementation for all four extraction strategies.
// It builds the FFmpeg command, runs it, parses the timestamp sidecar, computes
// perceptual hashes, and returns the candidate frame set.
//
// FFmpeg is run with its working directory set to the strategy subdirectory so
// that the output pattern and metadata sidecar can be plain relative filenames.
// This avoids Windows filter-graph escaping issues with drive-letter colons and
// backslashes in absolute paths.
func runStrategy(ctx context.Context, videoPath string, cfg Config, strategy Strategy, vf string) ([]candidateFrame, error) {
	dir := filepath.Join(cfg.WorkDir, string(strategy))
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("mkdir %s: %w", dir, err)
	}

	// Use plain relative paths — ffmpeg runs with Dir=dir so these resolve correctly.
	const metaRelPath = "meta.txt"
	const outRelPattern = "frame_%05d.png"

	// Append metadata sidecar and optional scale to the filter graph.
	fullVF := appendMetaAndScale(vf, metaRelPath, cfg.HashResizeWidth)

	cmd := iffmpeg.New(cfg.FFmpegPath,
		"-i", videoPath,
		"-vf", fullVF,
		"-vsync", "0",
		"-start_number", "0",
		outRelPattern,
	)
	cmd.Dir = dir

	if err := cmd.Run(ctx); err != nil {
		// Clean up partial output on failure so callers don't process corrupt files.
		os.RemoveAll(dir)
		return nil, fmt.Errorf("strategy %s ffmpeg: %w", strategy, err)
	}

	metaPath := filepath.Join(dir, metaRelPath)
	timestamps, err := iffmpeg.ParseMetadataFile(metaPath)
	if err != nil {
		// Non-fatal: sidecar may be absent if no frames were selected.
		timestamps = map[int]float64{}
	}

	return collectFrames(dir, strategy, timestamps, cfg)
}

// collectFrames walks the strategy output directory, computes a perceptual hash
// for each PNG, and returns the candidate set.
func collectFrames(dir string, strategy Strategy, timestamps map[int]float64, cfg Config) ([]candidateFrame, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("read dir %s: %w", dir, err)
	}

	var candidates []candidateFrame
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".png") {
			continue
		}
		// Parse the 0-based frame index from "frame_00042.png".
		stem := strings.TrimSuffix(e.Name(), ".png")
		parts := strings.Split(stem, "_")
		idx, err := strconv.Atoi(parts[len(parts)-1])
		if err != nil {
			continue
		}

		path := filepath.Join(dir, e.Name())
		ts := timestamps[idx]

		h, err := computeHash(path, cfg.HashAlgo)
		if err != nil {
			// Skip unreadable/corrupt frames rather than aborting the strategy.
			continue
		}

		candidates = append(candidates, candidateFrame{
			path:         path,
			timestampSec: ts,
			strategy:     strategy,
			hash:         h,
		})
	}
	return candidates, nil
}

// sceneFilter builds the -vf string for scene change detection.
// The select filter must come before any pts manipulation; do not reorder.
func sceneFilter(cfg Config) string {
	return fmt.Sprintf("select=gt(scene\\,%g)", cfg.SceneThreshold)
}

func keyframeFilter(_ Config) string {
	return "select=eq(pict_type\\,I)"
}

func temporalFilter(cfg Config) string {
	return fmt.Sprintf("fps=1/%g", cfg.TemporalInterval)
}

func mpdecimateFilter(cfg Config) string {
	return fmt.Sprintf("mpdecimate=hi=%d:lo=%d:frac=%g",
		cfg.MPDecimateHi, cfg.MPDecimateLo, cfg.MPDecimateFrac)
}

// appendMetaAndScale appends an optional scale step and the metadata=print
// sidecar to the given filter graph string. metaPath must be a plain relative
// filename (no path separators) so it needs no filter-graph escaping.
func appendMetaAndScale(vf, metaPath string, resizeWidth int) string {
	if resizeWidth > 0 {
		vf = fmt.Sprintf("%s,scale=%d:-1", vf, resizeWidth)
	}
	return fmt.Sprintf("%s,metadata=print:file=%s", vf, metaPath)
}
