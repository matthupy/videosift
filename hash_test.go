package videosift

import "testing"

func TestHammingDistance(t *testing.T) {
	cases := []struct {
		a, b uint64
		want int
	}{
		{0, 0, 0},
		{0, 1, 1},
		{0, 0xFFFFFFFFFFFFFFFF, 64},
		{0b1010, 0b0101, 4},
	}
	for _, tc := range cases {
		if got := hammingDistance(tc.a, tc.b); got != tc.want {
			t.Errorf("hammingDistance(%016x, %016x) = %d, want %d", tc.a, tc.b, got, tc.want)
		}
	}
}
