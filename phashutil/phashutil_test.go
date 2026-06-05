package phashutil

import "testing"

// TestHammingDistance exercises HammingDistance with a table of cases that
// cover identical hashes, bitwise-opposite hashes, and partial-difference
// cases with a known small number of differing bits.
func TestHammingDistance(t *testing.T) {
	cases := []struct {
		name string
		a, b uint64
		want int
	}{
		{
			name: "identical zero hashes",
			a:    0,
			b:    0,
			want: 0,
		},
		{
			name: "identical non-zero hashes",
			a:    0xDEADBEEFCAFEBABE,
			b:    0xDEADBEEFCAFEBABE,
			want: 0,
		},
		{
			name: "identical all-ones hashes",
			a:    ^uint64(0),
			b:    ^uint64(0),
			want: 0,
		},
		{
			name: "bitwise opposite of zero",
			a:    0,
			b:    ^uint64(0),
			want: 64,
		},
		{
			name: "bitwise opposite of non-zero pattern",
			a:    0xA5A5A5A5A5A5A5A5,
			b:    ^uint64(0xA5A5A5A5A5A5A5A5),
			want: 64,
		},
		{
			name: "differ by 1 bit (lowest)",
			a:    0,
			b:    1,
			want: 1,
		},
		{
			name: "differ by 1 bit (highest)",
			a:    0,
			b:    1 << 63,
			want: 1,
		},
		{
			name: "differ by 3 bits",
			a:    0,
			b:    0b1011, // bits 0, 1, 3 set
			want: 3,
		},
		{
			name: "differ by 8 bits (low byte flipped)",
			a:    0,
			b:    0xFF,
			want: 8,
		},
		{
			name: "differ by 8 bits across non-zero base",
			a:    0x00000000FFFFFFFF,
			b:    0x000000FFFFFFFFFF, // differs in bits 32..39
			want: 8,
		},
		{
			name: "symmetric: swap a and b",
			a:    0x000000FFFFFFFFFF,
			b:    0x00000000FFFFFFFF,
			want: 8,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got := HammingDistance(tc.a, tc.b)
			if got != tc.want {
				t.Fatalf("HammingDistance(%#x, %#x) = %d, want %d", tc.a, tc.b, got, tc.want)
			}
		})
	}
}

// TestSimilarity exercises Similarity with a table of cases that mirror the
// HammingDistance cases. Exact float comparisons are used because
// 1 - k/64 is exactly representable in IEEE-754 float64 for every integer
// k in [0, 64].
func TestSimilarity(t *testing.T) {
	cases := []struct {
		name string
		a, b uint64
		want float64
	}{
		{
			name: "identical zero hashes -> 1.0",
			a:    0,
			b:    0,
			want: 1.0,
		},
		{
			name: "identical non-zero hashes -> 1.0",
			a:    0xDEADBEEFCAFEBABE,
			b:    0xDEADBEEFCAFEBABE,
			want: 1.0,
		},
		{
			name: "identical all-ones hashes -> 1.0",
			a:    ^uint64(0),
			b:    ^uint64(0),
			want: 1.0,
		},
		{
			name: "bitwise opposite of zero -> 0.0",
			a:    0,
			b:    ^uint64(0),
			want: 0.0,
		},
		{
			name: "bitwise opposite of non-zero pattern -> 0.0",
			a:    0xA5A5A5A5A5A5A5A5,
			b:    ^uint64(0xA5A5A5A5A5A5A5A5),
			want: 0.0,
		},
		{
			name: "differ by 1 bit -> 63/64",
			a:    0,
			b:    1,
			want: 1.0 - 1.0/64.0,
		},
		{
			name: "differ by 1 bit (highest) -> 63/64",
			a:    0,
			b:    1 << 63,
			want: 1.0 - 1.0/64.0,
		},
		{
			name: "differ by 3 bits -> 61/64",
			a:    0,
			b:    0b1011,
			want: 1.0 - 3.0/64.0,
		},
		{
			name: "differ by 8 bits -> 56/64 = 0.875",
			a:    0,
			b:    0xFF,
			want: 0.875,
		},
		{
			name: "differ by 8 bits across non-zero base -> 0.875",
			a:    0x00000000FFFFFFFF,
			b:    0x000000FFFFFFFFFF,
			want: 0.875,
		},
		{
			name: "differ by 32 bits -> 0.5",
			a:    0,
			b:    0x00000000FFFFFFFF,
			want: 0.5,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got := Similarity(tc.a, tc.b)
			if got != tc.want {
				t.Fatalf("Similarity(%#x, %#x) = %v, want %v", tc.a, tc.b, got, tc.want)
			}
			if got < 0.0 || got > 1.0 {
				t.Fatalf("Similarity(%#x, %#x) = %v out of range [0.0, 1.0]", tc.a, tc.b, got)
			}
		})
	}
}

// TestSimilarityMatchesHammingDistance verifies the algebraic relationship
// Similarity(a, b) == 1 - HammingDistance(a, b)/64 across a small spread of
// hand-picked inputs. This guards against the two functions drifting out of
// sync if someone edits one without the other.
func TestSimilarityMatchesHammingDistance(t *testing.T) {
	cases := []struct {
		name string
		a, b uint64
	}{
		{"both zero", 0, 0},
		{"zero vs all-ones", 0, ^uint64(0)},
		{"low byte", 0, 0xFF},
		{"alternating pattern", 0xAAAAAAAAAAAAAAAA, 0x5555555555555555},
		{"arbitrary pair", 0x0123456789ABCDEF, 0xFEDCBA9876543210},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			d := HammingDistance(tc.a, tc.b)
			want := 1.0 - float64(d)/64.0
			got := Similarity(tc.a, tc.b)
			if got != want {
				t.Fatalf("Similarity(%#x, %#x) = %v, want %v (HammingDistance = %d)", tc.a, tc.b, got, want, d)
			}
		})
	}
}
