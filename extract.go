package videosift

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	iffmpeg "github.com/matthupy/videosift/internal/ffmpeg"
	"golang.org/x/sync/errgroup"
)

// Extract runs the full extraction pipeline for videoPath and returns the
// deduplicated, ordered set of frames.
//
// When cfg.WorkDir is empty, Extract creates a temporary directory for
// intermediate files and removes it before returning, leaving only the
// final []Frame paths valid for the lifetime of the returned slice — callers
// must copy the files if they need them after Extract returns. When WorkDir
// is set by the caller, the extracted PNGs persist and the caller is
// responsible for cleanup.
func Extract(ctx context.Context, videoPath string, cfg Config) ([]Frame, error) {
	applyDefaults(&cfg)

	if err := validateBinaries(&cfg); err != nil {
		return nil, err
	}

	if !anyEnabled(cfg) {
		return nil, ErrNoStrategies
	}

	info, err := probe(ctx, videoPath, cfg.FFprobePath)
	if err != nil {
		return nil, fmt.Errorf("probe: %w", err)
	}

	// Manage WorkDir lifetime.
	callerOwnsDir := cfg.WorkDir != ""
	if !callerOwnsDir {
		tmp, err := os.MkdirTemp("", "videosift-*")
		if err != nil {
			return nil, fmt.Errorf("create temp dir: %w", err)
		}
		defer os.RemoveAll(tmp)
		cfg.WorkDir = tmp
	}

	candidates, err := runStrategies(ctx, videoPath, cfg)
	if err != nil {
		return nil, err
	}
	if len(candidates) == 0 {
		return nil, ErrNoFrames
	}

	candidates = merge(candidates)
	candidates = hammingDedup(candidates, cfg.HammingThreshold)
	candidates = capFrames(candidates, cfg.MaxFrames)

	if len(candidates) == 0 {
		return nil, ErrNoFrames
	}

	frames, err := finalizeFrames(candidates, cfg.WorkDir, &info, callerOwnsDir)
	if err != nil {
		return nil, err
	}
	return frames, nil
}

// applyDefaults fills zero-value Config fields with sensible defaults.
func applyDefaults(cfg *Config) {
	if cfg.FFmpegPath == "" {
		cfg.FFmpegPath = "ffmpeg"
	}
	if cfg.FFprobePath == "" {
		cfg.FFprobePath = "ffprobe"
	}
	if cfg.SceneThreshold == 0 {
		cfg.SceneThreshold = 0.4
	}
	if cfg.TemporalInterval == 0 {
		cfg.TemporalInterval = 2.0
	}
	if cfg.MPDecimateHi == 0 {
		cfg.MPDecimateHi = 768
	}
	if cfg.MPDecimateLo == 0 {
		cfg.MPDecimateLo = 320
	}
	if cfg.MPDecimateFrac == 0 {
		cfg.MPDecimateFrac = 0.33
	}
	if cfg.HashAlgo == "" {
		cfg.HashAlgo = HashPHash
	}
	if cfg.HammingThreshold == 0 {
		cfg.HammingThreshold = 8
	}
	if cfg.HashResizeWidth == 0 {
		cfg.HashResizeWidth = 256
	}
}

func validateBinaries(cfg *Config) error {
	ffmpeg, err := iffmpeg.LookPath(cfg.FFmpegPath)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrNoBinaries, err)
	}
	ffprobe, err := iffmpeg.LookPath(cfg.FFprobePath)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrNoBinaries, err)
	}
	cfg.FFmpegPath = ffmpeg
	cfg.FFprobePath = ffprobe
	return nil
}

func anyEnabled(cfg Config) bool {
	return cfg.Scene || cfg.Keyframe || cfg.Temporal || cfg.MPDecimate
}

type strategyFunc func(context.Context, string, Config) ([]candidateFrame, error)

// runStrategies fans out to all enabled extraction strategies concurrently,
// limited by cfg.Concurrency. If any strategy fails, all in-flight strategies
// are cancelled and the error is returned.
func runStrategies(ctx context.Context, videoPath string, cfg Config) ([]candidateFrame, error) {
	type task struct {
		name string
		fn   strategyFunc
	}

	var tasks []task
	if cfg.Scene {
		tasks = append(tasks, task{"scene", runScene})
	}
	if cfg.Keyframe {
		tasks = append(tasks, task{"keyframe", runKeyframe})
	}
	if cfg.Temporal {
		tasks = append(tasks, task{"temporal", runTemporal})
	}
	if cfg.MPDecimate {
		tasks = append(tasks, task{"mpdecimate", runMPDecimate})
	}

	concurrency := cfg.Concurrency
	if concurrency <= 0 {
		concurrency = len(tasks)
	}

	sem := make(chan struct{}, concurrency)
	var mu sync.Mutex
	var all []candidateFrame

	g, gctx := errgroup.WithContext(ctx)
	for _, t := range tasks {
		t := t
		g.Go(func() error {
			sem <- struct{}{}
			defer func() { <-sem }()

			frames, err := t.fn(gctx, videoPath, cfg)
			if err != nil {
				return fmt.Errorf("strategy %s: %w", t.name, err)
			}
			mu.Lock()
			all = append(all, frames...)
			mu.Unlock()
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}
	return all, nil
}

// finalizeFrames renames surviving candidates to sequentially numbered files
// in WorkDir, deletes unselected PNGs, and constructs the public Frame slice.
func finalizeFrames(candidates []candidateFrame, workDir string, info *VideoInfo, keepInPlace bool) ([]Frame, error) {
	selectedPaths := make(map[string]bool, len(candidates))
	frames := make([]Frame, len(candidates))

	for i, c := range candidates {
		newName := fmt.Sprintf("frame_%05d.png", i)
		newPath := filepath.Join(workDir, newName)

		if c.path != newPath {
			if err := os.Rename(c.path, newPath); err != nil {
				// Rename across volumes (e.g. different strategy subdirs on Windows)
				// shouldn't occur since all subdirs share WorkDir, but handle gracefully.
				return nil, fmt.Errorf("rename frame %d: %w", i, err)
			}
		}
		selectedPaths[newPath] = true
		frames[i] = Frame{
			Index:        i,
			TimestampSec: c.timestampSec,
			Strategy:     c.strategy,
			Path:         newPath,
			Hash:         c.hash,
			Video:        info,
		}
	}

	// Remove unselected PNGs from strategy subdirectories.
	if err := cleanUnselected(workDir, selectedPaths); err != nil {
		// Non-fatal: leftover files in a temp dir are cleaned up by the deferred RemoveAll.
		_ = err
	}
	return frames, nil
}

// cleanUnselected removes all PNG files under workDir that are not in the
// selected set, including files inside strategy subdirectories.
func cleanUnselected(workDir string, keep map[string]bool) error {
	entries, err := os.ReadDir(workDir)
	if err != nil {
		return err
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		subDir := filepath.Join(workDir, e.Name())
		subEntries, err := os.ReadDir(subDir)
		if err != nil {
			continue
		}
		for _, se := range subEntries {
			if se.IsDir() {
				continue
			}
			p := filepath.Join(subDir, se.Name())
			if !keep[p] {
				os.Remove(p)
			}
		}
		// Remove the now-empty strategy subdirectory.
		os.Remove(subDir)
	}
	return nil
}
