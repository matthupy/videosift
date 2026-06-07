package phashutil

import (
	"math"
	"testing"
)

func TestHammingDistanceAndSimilarity(t *testing.T) {
	tests := []struct {
		name      string
		hash1     uint64
		hash2     uint64
		wantDist  int
		wantSim   float64
	}{
		{
			name:     "identical hashes",
			hash1:    0x1234567890ABCDEF,
			hash2:    0x1234567890ABCDEF,
			wantDist: 0,
			wantSim:  1.0,
		},
		{
			name:     "bitwise-opposite hashes",
			hash1:    0x0000000000000000,
			hash2:    ^uint64(0),
			wantDist: 64,
			wantSim:  0.0,
		},
		{
			name:     "partial mismatch (5 bits differ)",
			hash1:    0x0000000000000000,
			hash2:    0x000000000000001F, // Popcount is 5 (binary 11111)
			wantDist: 5,
			wantSim:  0.921875, // (64 - 5) / 64
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotDist := HammingDistance(tt.hash1, tt.hash2)
			if gotDist != tt.wantDist {
				t.Errorf("HammingDistance(%x, %x) = %d; want %d", tt.hash1, tt.hash2, gotDist, tt.wantDist)
			}

			gotSim := Similarity(tt.hash1, tt.hash2)
			if math.Abs(gotSim-tt.wantSim) > 1e-9 {
				t.Errorf("Similarity(%x, %x) = %.4f; want %.4f", tt.hash1, tt.hash2, gotSim, tt.wantSim)
			}
		})
	}
}
