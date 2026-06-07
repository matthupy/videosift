package phashutil

import (
	"testing"
)

func TestHammingDistanceAndSimilarity(t *testing.T) {
	tests := []struct {
		name     string
		hashA    uint64
		hashB    uint64
		wantDist int
		wantSim  float64
	}{
		{
			name:     "identical hashes",
			hashA:    0x123456789ABCDEF,
			hashB:    0x123456789ABCDEF,
			wantDist: 0,
			wantSim:  1.0,
		},
		{
			name:     "bitwise-opposite hashes",
			hashA:    uint64(0xFFFFFFFFFFFFFFFF),
			hashB:    uint64(0x0000000000000000),
			wantDist: 64,
			wantSim:  0.0,
		},
		{
			name:     "partial match with exactly 32 differing bits",
			hashA:    0xFFFF_FFFF_FFFF_FFFF,
			hashB:    0xAAAA_AAAA_AAAA_AAAA,
			wantDist: 32,
			wantSim:  0.5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dist := HammingDistance(tt.hashA, tt.hashB)
			if dist != tt.wantDist {
				t.Errorf("HammingDistance() = %d, want %d", dist, tt.wantDist)
			}

			sim := Similarity(tt.hashA, tt.hashB)
			if sim != tt.wantSim {
				t.Errorf("Similarity() = %v, want %v", sim, tt.wantSim)
			}
		})
	}
}