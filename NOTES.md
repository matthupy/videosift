Implementation decisions and trade-offs:

1. Same package vs _test package: Per the task, used `package phashutil` (not `phashutil_test`) so the tests can call HammingDistance and Similarity directly without an import path. This matches the task's 'keep it simple and self-contained' guidance.

2. Exact float comparison: The task explicitly authorizes exact `==` comparisons for Similarity because 1 - k/64 is exactly representable in float64 for every integer k in [0, 64] (64 is 2^6, so k/64 has at most 6 fractional bits — far inside the 52-bit mantissa). I took advantage of this rather than introducing an epsilon, which would have weakened the test.

3. Coverage choices: Three distinct test functions instead of one mega-table:
   - TestHammingDistance and TestSimilarity each focus on one function so failures point at the right symbol.
   - TestSimilarityMatchesHammingDistance is a small invariant test that locks in the algebraic relationship between the two functions. This is cheap insurance and would catch refactors that touch only one of them.

4. Inputs were chosen to be informative:
   - 0xDEADBEEFCAFEBABE — non-zero, non-symmetric pattern for the 'identical non-zero' case.
   - 0xA5A5A5A5A5A5A5A5 and its bitwise complement — exercises opposites on a non-trivial pattern (not just 0 vs all-ones).
   - 0xAAAA…/0x5555… (alternating bits) — every bit differs, exercises POPCNT on a regular pattern.
   - Bit 0 and bit 63 single-bit differences — covers both ends of the word in case there's ever a byte-order or shift bug.
   - 0b1011 — gives exactly 3 set bits with an asymmetric pattern (not just three consecutive low bits).

5. Subtest hygiene: Used the standard `tc := tc` capture before `t.Run` so the loop variable is safely bound per subtest. This is the idiomatic guard and is harmless even on Go 1.22+ where the loopvar change makes it redundant — it keeps the tests robust if anyone later parallelizes them with `t.Parallel()`.

6. Failure messages: Used `%#x` for hash values so failing outputs are immediately readable as hex, and included both got and want plus the inputs in every failure message.

7. No external dependencies: Only `testing` is imported. No table helpers, no testify, nothing beyond the standard library — matches the task and the implementation file's constraints.

8. Reviewer attention: The deliberate use of exact float `==` is the one thing a reviewer might flinch at. It is correct here (see point 2), and I'd rather assert exactness than introduce a tolerance that could hide a real bug in Similarity. If the project later adds non-integer-bit-count metrics, those tests should switch to an epsilon comparison — but for the current API it would be the wrong call.