package videosift

import (
	"testing"
)

func TestMerge_sortsByTimestamp(t *testing.T) {
	in := []candidateFrame{
		{timestampSec: 3.0, strategy: StrategyTemporal},
		{timestampSec: 1.0, strategy: StrategyScene},
		{timestampSec: 2.0, strategy: StrategyKeyframe},
	}
	out := merge(in)
	want := []float64{1.0, 2.0, 3.0}
	for i, f := range out {
		if f.timestampSec != want[i] {
			t.Errorf("position %d: got %.1f want %.1f", i, f.timestampSec, want[i])
		}
	}
}

func TestMerge_collapsesSameTimestamp(t *testing.T) {
	in := []candidateFrame{
		{timestampSec: 1.0, strategy: StrategyTemporal},
		{timestampSec: 1.0, strategy: StrategyScene},    // lower precedence — should win
		{timestampSec: 1.0, strategy: StrategyKeyframe},
	}
	out := merge(in)
	if len(out) != 1 {
		t.Fatalf("want 1 frame, got %d", len(out))
	}
	if out[0].strategy != StrategyScene {
		t.Errorf("want strategy=scene, got %s", out[0].strategy)
	}
}

func TestHammingDedup_removesNearDuplicates(t *testing.T) {
	// hash=0 and hash=1 differ by 1 bit → duplicate under threshold=8.
	// hash=0 and hash=0xFFFF differ by 16 bits → not duplicate.
	in := []candidateFrame{
		{timestampSec: 0.0, hash: 0x0000},
		{timestampSec: 1.0, hash: 0x0001}, // 1-bit diff from first → dropped
		{timestampSec: 2.0, hash: 0xFFFF}, // 16-bit diff → kept
	}
	out := hammingDedup(in, 8)
	if len(out) != 2 {
		t.Fatalf("want 2 frames after dedup, got %d", len(out))
	}
	if out[0].hash != 0x0000 || out[1].hash != 0xFFFF {
		t.Errorf("unexpected hashes: %x %x", out[0].hash, out[1].hash)
	}
}

func TestHammingDedup_zeroThresholdKeepsAll(t *testing.T) {
	in := []candidateFrame{
		{hash: 0x0000},
		{hash: 0x0001},
		{hash: 0x0002},
	}
	out := hammingDedup(in, 0)
	if len(out) != 3 {
		t.Errorf("want 3 frames with threshold=0, got %d", len(out))
	}
}

func TestCapFrames(t *testing.T) {
	frames := make([]candidateFrame, 10)
	for i := range frames {
		frames[i].timestampSec = float64(i)
	}

	out := capFrames(frames, 5)
	if len(out) != 5 {
		t.Fatalf("want 5 frames, got %d", len(out))
	}
	// First and last must be included.
	if out[0].timestampSec != 0 {
		t.Errorf("first frame should be ts=0, got %.1f", out[0].timestampSec)
	}
	if out[len(out)-1].timestampSec != 9 {
		t.Errorf("last frame should be ts=9, got %.1f", out[len(out)-1].timestampSec)
	}
}

func TestCapFrames_noCapWhenUnderMax(t *testing.T) {
	frames := make([]candidateFrame, 3)
	out := capFrames(frames, 10)
	if len(out) != 3 {
		t.Errorf("want 3 (unchanged), got %d", len(out))
	}
}

func TestCapFrames_zeroMaxNoOp(t *testing.T) {
	frames := make([]candidateFrame, 5)
	out := capFrames(frames, 0)
	if len(out) != 5 {
		t.Errorf("want 5 (unchanged), got %d", len(out))
	}
}
