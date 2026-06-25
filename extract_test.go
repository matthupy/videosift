package videosift

import (
	"strings"
	"testing"
)

// TestApplyDefaults_preservesDisableSentinels verifies the documented contract
// that an explicit HammingThreshold == 0 (disable dedup) and HashResizeWidth == 0
// (disable rescale) survive applyDefaults rather than being clobbered to 8/256.
func TestApplyDefaults_preservesDisableSentinels(t *testing.T) {
	cfg := Config{HammingThreshold: 0, HashResizeWidth: 0}
	applyDefaults(&cfg)

	if cfg.HammingThreshold != 0 {
		t.Errorf("HammingThreshold: explicit 0 should stick (dedup disabled), got %d", cfg.HammingThreshold)
	}
	if cfg.HashResizeWidth != 0 {
		t.Errorf("HashResizeWidth: explicit 0 should stick (rescale disabled), got %d", cfg.HashResizeWidth)
	}
}

// TestApplyDefaults_leavesNonZeroValuesUntouched confirms explicit non-zero
// values are not overridden.
func TestApplyDefaults_leavesNonZeroValuesUntouched(t *testing.T) {
	cfg := Config{HammingThreshold: 12, HashResizeWidth: 512}
	applyDefaults(&cfg)

	if cfg.HammingThreshold != 12 {
		t.Errorf("HammingThreshold: want 12, got %d", cfg.HammingThreshold)
	}
	if cfg.HashResizeWidth != 512 {
		t.Errorf("HashResizeWidth: want 512, got %d", cfg.HashResizeWidth)
	}
}

// TestApplyDefaults_otherFieldsStillDefaulted ensures the unrelated clobbers
// (left intact by the fix) keep working from a zero-value Config.
func TestApplyDefaults_otherFieldsStillDefaulted(t *testing.T) {
	cfg := Config{}
	applyDefaults(&cfg)

	if cfg.SceneThreshold != 0.4 {
		t.Errorf("SceneThreshold: want 0.4, got %v", cfg.SceneThreshold)
	}
	if cfg.TemporalInterval != 2.0 {
		t.Errorf("TemporalInterval: want 2.0, got %v", cfg.TemporalInterval)
	}
	if cfg.HashAlgo != HashPHash {
		t.Errorf("HashAlgo: want phash, got %q", cfg.HashAlgo)
	}
}

// TestExtractDedupDisabled_keepsAllCandidates proves that with the disable
// sentinel (HammingThreshold == 0) the dedup pass — reached via the public
// Extract path's value — keeps every candidate.
func TestExtractDedupDisabled_keepsAllCandidates(t *testing.T) {
	cfg := Config{HammingThreshold: 0}
	applyDefaults(&cfg)

	// Three near-identical hashes that WOULD collapse under the default threshold.
	in := []candidateFrame{
		{hash: 0x0000},
		{hash: 0x0001},
		{hash: 0x0003},
	}
	out := hammingDedup(in, cfg.HammingThreshold)
	if len(out) != 3 {
		t.Fatalf("dedup disabled: want all 3 candidates kept, got %d", len(out))
	}
}

// TestAppendMetaAndScale_zeroWidthSkipsRescale verifies HashResizeWidth == 0
// produces no scale step in the FFmpeg filter graph.
func TestAppendMetaAndScale_zeroWidthSkipsRescale(t *testing.T) {
	got := appendMetaAndScale("select='eq(pict_type,I)'", "meta.txt", 0)
	if strings.Contains(got, "scale=") {
		t.Errorf("width 0 should skip rescale, but filter contains scale: %q", got)
	}
}

// TestAppendMetaAndScale_nonZeroWidthAddsScale verifies a positive width injects
// the scale=W:-1 step.
func TestAppendMetaAndScale_nonZeroWidthAddsScale(t *testing.T) {
	got := appendMetaAndScale("select='eq(pict_type,I)'", "meta.txt", 256)
	if !strings.Contains(got, "scale=256:-1") {
		t.Errorf("width 256 should add scale=256:-1, got %q", got)
	}
}
