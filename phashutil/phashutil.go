// Package phashutil provides utilities for comparing 64-bit perceptual hashes.
//
// It exposes functions to calculate the Hamming distance and normalized similarity
// between two hashes, suitable for applications requiring fast hash comparison
// without external dependencies.
package phashutil

import (
	"math/bits"
)

// HammingDistance returns the count of differing bits between two 64-bit values.
// It is optimized to work efficiently on all possible uint64 inputs by leveraging
// bitwise XOR and population counting.
func HammingDistance(a, b uint64) int {
	return bits.OnesCount64(a ^ b)
}

// Similarity calculates the perceptual-hash similarity between two 64-bit values
// as a floating-point value in the range [0.0, 1.0].
//
// The formula used is: 1.0 - float64(HammingDistance(a,b))/64.0.
// A result of 1.0 indicates identical hashes, while 0.0 indicates completely
// inverted bits. Values in between represent the degree of overlap.
func Similarity(a, b uint64) float64 {
	return 1.0 - float64(HammingDistance(a,b))/64.0
}