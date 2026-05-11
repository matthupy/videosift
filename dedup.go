package videosift

import "sort"

// merge combines all candidate slices, sorts by (timestampSec ASC,
// strategyPrecedence ASC), then collapses exact-timestamp duplicates by
// keeping the frame with the lowest strategy precedence.
func merge(candidates []candidateFrame) []candidateFrame {
	sort.Slice(candidates, func(i, j int) bool {
		a, b := candidates[i], candidates[j]
		if a.timestampSec != b.timestampSec {
			return a.timestampSec < b.timestampSec
		}
		return a.strategy.precedence() < b.strategy.precedence()
	})

	// Collapse exact-timestamp duplicates: keep the first (lowest precedence).
	deduped := candidates[:0]
	for i, c := range candidates {
		if i == 0 || c.timestampSec != deduped[len(deduped)-1].timestampSec {
			deduped = append(deduped, c)
		}
	}
	return deduped
}

// hammingDedup walks candidates in order and accepts a frame only when its
// Hamming distance to every already-accepted frame exceeds threshold.
// When threshold == 0, all candidates are accepted (dedup disabled).
func hammingDedup(candidates []candidateFrame, threshold int) []candidateFrame {
	if threshold == 0 {
		return candidates
	}

	accepted := make([]candidateFrame, 0, len(candidates))
	for _, c := range candidates {
		if !tooSimilar(c.hash, accepted, threshold) {
			accepted = append(accepted, c)
		}
	}
	return accepted
}

func tooSimilar(hash uint64, accepted []candidateFrame, threshold int) bool {
	for _, a := range accepted {
		if hammingDistance(hash, a.hash) <= threshold {
			return true
		}
	}
	return false
}

// capFrames reduces the frame set to at most max entries by uniform stride,
// always including the first and last frame to preserve boundary coverage.
// If max <= 0 or max >= len(frames), the input is returned unchanged.
func capFrames(frames []candidateFrame, max int) []candidateFrame {
	if max <= 0 || len(frames) <= max {
		return frames
	}
	if max == 1 {
		return frames[:1]
	}

	selected := make([]candidateFrame, 0, max)
	// Always include first and last; distribute the remaining (max-2) evenly.
	step := float64(len(frames)-1) / float64(max-1)
	for i := 0; i < max; i++ {
		idx := int(float64(i)*step + 0.5)
		if idx >= len(frames) {
			idx = len(frames) - 1
		}
		selected = append(selected, frames[idx])
	}
	return selected
}
