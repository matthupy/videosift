package phashutil

import (
	"testing"
)

func TestHammingDistance(t *testing.T) {
	tests := []struct {
		name string
		a    uint64
		b    uint64
		want int
	}{
		{
			name: "identical hashes",
			a:    0x123456789ABCDEF0,
			b:    0x123456789ABCDEF0,
			want: 0,
		},
		{
			name: "bitwise-opposite hashes",
			a:    uint64(0x123456789ABCDEF0),
			b:    ^uint64(0x123456789ABCDEF0),
			want: 64,
		},
		{
			name: "non-trivial partial case",
			a:    uint64(0xFFFFFFFFFFFFFFFF),
			b:    uint64(0xFFFFFFFFFFFFFFFE),
			want: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := HammingDistance(tt.a, tt.b)
			if got != tt.want {
				t.Errorf("HammingDistance() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSimilarity(t *testing.T) {
	tests := []struct {
		name string
		a    uint64
		b    uint64
		want float64
	}{
		{
			name: "identical hashes",
			a:    0x123456789ABCDEF0,
			b:    0x123456789ABCDEF0,
			want: 1.0,
		},
		{
			name: "bitwise-opposite hashes",
			a:    uint64(0x123456789ABCDEF0),
			b:    ^uint64(0x123456789ABCDEF0),
			want: 0.0,
		},
		{
			name: "non-trivial partial case",
			a:    uint64(0xFFFFFFFFFFFFFFFF),
			b:    uint64(0xFFFFFFFFFFFFFFFE),
			want: 0.984375,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Similarity(tt.a, tt.b)
			if got != tt.want {
				t.Errorf("Similarity() = %v, want %v", got, tt.want)
			}
		})
	}
}