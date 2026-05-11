package ffmpeg

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseRationalFPS(t *testing.T) {
	cases := []struct {
		input   string
		want    float64
		wantErr bool
	}{
		{"30/1", 30.0, false},
		{"30000/1001", 29.97002997002997, false},
		{"25/1", 25.0, false},
		{"0/0", 0, true},
		{"bad", 0, true},
		{"30", 0, true},
	}
	for _, tc := range cases {
		got, err := ParseRationalFPS(tc.input)
		if (err != nil) != tc.wantErr {
			t.Errorf("ParseRationalFPS(%q) error=%v wantErr=%v", tc.input, err, tc.wantErr)
			continue
		}
		if !tc.wantErr && abs(got-tc.want) > 1e-9 {
			t.Errorf("ParseRationalFPS(%q) = %v, want %v", tc.input, got, tc.want)
		}
	}
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

func TestParseMetadataFile(t *testing.T) {
	content := `frame:0  pts:0  pts_time:0.000
lavfi.scene_score=0.000

frame:1  pts:512  pts_time:0.512
lavfi.scene_score=0.612

frame:2  pts:1024  pts_time:1.024
lavfi.scene_score=0.215
`
	tmp := filepath.Join(t.TempDir(), "meta.txt")
	if err := os.WriteFile(tmp, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	got, err := ParseMetadataFile(tmp)
	if err != nil {
		t.Fatalf("ParseMetadataFile: %v", err)
	}

	want := map[int]float64{0: 0.0, 1: 0.512, 2: 1.024}
	for k, v := range want {
		if got[k] != v {
			t.Errorf("frame %d: got %v want %v", k, got[k], v)
		}
	}
	if len(got) != len(want) {
		t.Errorf("got %d entries, want %d", len(got), len(want))
	}
}

func TestParseMetadataFileMissing(t *testing.T) {
	_, err := ParseMetadataFile(filepath.Join(t.TempDir(), "nonexistent.txt"))
	if err == nil {
		t.Error("expected error for missing file")
	}
}
