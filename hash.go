package videosift

import (
	"fmt"
	"image"
	_ "image/png"
	"math/bits"
	"os"

	"github.com/corona10/goimagehash"
)

// computeHash opens a PNG file and returns its perceptual hash using the
// algorithm specified in algo.
func computeHash(path string, algo HashAlgo) (uint64, error) {
	f, err := os.Open(path)
	if err != nil {
		return 0, fmt.Errorf("open %s: %w", path, err)
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	if err != nil {
		return 0, fmt.Errorf("decode %s: %w", path, err)
	}

	switch algo {
	case HashDHash:
		h, err := goimagehash.DifferenceHash(img)
		if err != nil {
			return 0, err
		}
		return h.GetHash(), nil
	default: // HashPHash
		h, err := goimagehash.PerceptionHash(img)
		if err != nil {
			return 0, err
		}
		return h.GetHash(), nil
	}
}

// hammingDistance returns the number of differing bits between two hash values.
func hammingDistance(a, b uint64) int {
	return bits.OnesCount64(a ^ b)
}
