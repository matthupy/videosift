package videosift

// Strategy identifies which FFmpeg extraction method surfaced a frame.
type Strategy string

const (
	StrategyScene      Strategy = "scene"
	StrategyKeyframe   Strategy = "keyframe"
	StrategyTemporal   Strategy = "temporal"
	StrategyMPDecimate Strategy = "mpdecimate"
)

// strategyPrecedence returns a sort key for tie-breaking when multiple
// strategies produce a frame at the same timestamp. Lower wins.
func (s Strategy) precedence() int {
	switch s {
	case StrategyScene:
		return 1
	case StrategyKeyframe:
		return 2
	case StrategyTemporal:
		return 3
	case StrategyMPDecimate:
		return 4
	}
	return 99
}

// HashAlgo selects the perceptual hash function used for deduplication.
type HashAlgo string

const (
	HashPHash HashAlgo = "phash"
	HashDHash HashAlgo = "dhash"
)

// Config controls every tunable parameter of the extraction pipeline.
// Use DefaultConfig() to obtain sensible defaults, then override as needed.
//
// Note the zero-value asymmetry: applyDefaults fills in every zero-valued
// field except HammingThreshold and HashResizeWidth, whose zero value means
// "disabled" rather than "use default". A from-scratch Config{} therefore
// runs with hash dedup and resize off; use DefaultConfig() to get 8/256.
type Config struct {
	// Strategy toggles — all enabled by default.
	Scene      bool
	Keyframe   bool
	Temporal   bool
	MPDecimate bool

	// SceneThreshold is the scene-change score [0.0, 1.0] that triggers
	// frame selection. Lower values are more sensitive. Default: 0.4.
	SceneThreshold float64

	// TemporalInterval is the number of seconds between uniformly sampled
	// frames (i.e. fps=1/TemporalInterval). Default: 2.0.
	TemporalInterval float64

	// MPDecimate filter parameters. Defaults map to FFmpeg's built-in defaults.
	MPDecimateHi   int     // default 768  (64*12)
	MPDecimateLo   int     // default 320  (64*5)
	MPDecimateFrac float64 // default 0.33

	// HashAlgo selects pHash or dHash for Hamming-distance deduplication.
	// Default: HashPHash.
	HashAlgo HashAlgo

	// HammingThreshold is the maximum Hamming distance at which two frames
	// are considered duplicates; the later frame is dropped. Default: 8.
	// Set to 0 to disable hash-based deduplication.
	HammingThreshold int

	// HashResizeWidth is the pixel width frames are scaled to before hashing.
	// Smaller values speed up hashing with minimal accuracy loss. Default: 256.
	// Set to 0 to disable rescaling (not recommended for large videos).
	HashResizeWidth int

	// MaxFrames caps the size of the final frame set. When the deduplicated
	// count exceeds MaxFrames, frames are selected by uniform stride to
	// preserve temporal spread. 0 means no cap.
	MaxFrames int

	// WorkDir is the directory where extracted PNGs are written.
	// If empty, Extract creates a temporary directory and removes it on return.
	// If set, the caller owns the directory and is responsible for cleanup.
	WorkDir string

	// FFmpegPath and FFprobePath override binary discovery via PATH.
	FFmpegPath  string
	FFprobePath string

	// Concurrency limits how many strategy goroutines run simultaneously.
	// Defaults to the number of enabled strategies (all run in parallel).
	Concurrency int
}

// DefaultConfig returns a Config with all strategies enabled and sensible
// defaults suitable for most content-moderation workloads.
func DefaultConfig() Config {
	return Config{
		Scene:            true,
		Keyframe:         true,
		Temporal:         true,
		MPDecimate:       true,
		SceneThreshold:   0.4,
		TemporalInterval: 2.0,
		MPDecimateHi:     768,
		MPDecimateLo:     320,
		MPDecimateFrac:   0.33,
		HashAlgo:         HashPHash,
		HammingThreshold: 8,
		HashResizeWidth:  256,
		FFmpegPath:       "ffmpeg",
		FFprobePath:      "ffprobe",
	}
}

// VideoInfo holds source video metadata populated by ffprobe.
type VideoInfo struct {
	Duration  float64 // total duration in seconds
	Width     int
	Height    int
	Codec     string  // e.g. "h264"
	FrameRate float64 // average frames per second
	BitRate   int64   // bits per second
}

// Frame is one extracted, deduplicated frame in the final result set.
type Frame struct {
	// Index is the 0-based position in the final ordered set.
	Index int

	// TimestampSec is the presentation timestamp of this frame in seconds.
	TimestampSec float64

	// Strategy is the extraction method that first produced this frame.
	Strategy Strategy

	// Path is the absolute path to the extracted PNG file.
	Path string

	// Hash is the perceptual hash value, available for downstream use.
	Hash uint64

	// Video holds source video metadata. The same pointer is shared across
	// all frames in a result set.
	Video *VideoInfo
}

// candidateFrame is the internal representation before final selection.
type candidateFrame struct {
	path         string
	timestampSec float64
	strategy     Strategy
	hash         uint64
}
