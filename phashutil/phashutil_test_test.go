All test code lives in phashutil/phashutil_test.go (see files[0] above). Coverage summary:

- TestHammingDistance: 11 table-driven subtests covering
  * identical hashes (zero, non-zero pattern 0xDEADBEEFCAFEBABE, all-ones)
  * bitwise-opposite hashes (from zero and from a non-zero pattern 0xA5A5A5A5A5A5A5A5) — both expect distance 64
  * partial-difference cases: 1 bit (lowest and highest position), 3 bits (0b1011), and 8 bits (low byte 0xFF, and a non-zero-base case differing in bits 32..39)
  * symmetry check: swapping a and b yields the same distance

- TestSimilarity: 11 table-driven subtests mirroring the HammingDistance cases and asserting exact float64 equality (justified because 1 - k/64 for integer k in [0,64] is exactly representable in IEEE-754), plus a range guard that the result is in [0.0, 1.0]. Includes the spec'd checks Similarity(x, x) == 1.0 (for x = 0 and x = 0xDEADBEEFCAFEBABE), Similarity(a, ^a) == 0.0, and partial cases with known similarity values (e.g. 1 - 1/64, 1 - 3/64, 0.875, 0.5).

- TestSimilarityMatchesHammingDistance: 5 cross-check subtests verifying the algebraic identity Similarity(a, b) == 1 - HammingDistance(a, b)/64 across hand-picked inputs (both zero, zero vs all-ones, low byte, alternating 0xAAAA…/0x5555…, and an arbitrary pair). This guards against the two functions drifting apart in future edits.

All tests use only the standard `testing` package, are table-driven with descriptive `name` fields, and use `t.Run(name, ...)` subtests so failures pinpoint the exact case.