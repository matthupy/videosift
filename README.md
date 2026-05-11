# videosift

A Go library and CLI tool that extracts the minimum set of visually unique frames from a video file, optimized for use with image-based machine learning models such as content moderation classifiers.

## How it works

videosift runs up to four FFmpeg-based extraction strategies in parallel, then deduplicates the full candidate set using perceptual hashing and Hamming distance comparison:

1. **Scene change detection** — selects frames where the visual content changes significantly
2. **Keyframe extraction** — pulls only the video's encoded I-frames (fast, no full decode)
3. **Uniform temporal sampling** — one frame every N seconds as a coverage baseline
4. **MPDecimate** — drops near-duplicate consecutive frames during decode

The result is an ordered, deduplicated slice of `Frame` values with timestamps, extraction method labels, and source video metadata attached.

## Prerequisites

FFmpeg and FFprobe must be installed and available on `PATH`. No CGo or other native dependencies are required.

**macOS:** `brew install ffmpeg`  
**Ubuntu/Debian:** `apt install ffmpeg`  
**Windows:** Download from [ffmpeg.org](https://ffmpeg.org/download.html) and add to PATH.

## Installation

```bash
go get github.com/matthupy/videosift
```

CLI tool:

```bash
go install github.com/matthupy/videosift/cmd/extract@latest
```

## Library usage

```go
import "github.com/matthupy/videosift"

cfg := videosift.DefaultConfig()
// cfg.MaxFrames = 30        // cap output
// cfg.SceneThreshold = 0.3  // more sensitive scene detection
// cfg.WorkDir = "./frames"   // persist output (caller cleans up)

frames, err := videosift.Extract(ctx, "input.mp4", cfg)
if err != nil {
    log.Fatal(err)
}

for _, f := range frames {
    fmt.Printf("[%02d] %.3fs  %-12s  %s\n", f.Index, f.TimestampSec, f.Strategy, f.Path)
}
```

When `WorkDir` is left empty, extracted PNGs are written to a temporary directory that is removed before `Extract` returns — copy the files in the returned `[]Frame` if you need them to persist.

## CLI usage

```
extract -i input.mp4 [options]

Options:
  -i string          Input video file (required)
  -o string          Output directory for extracted PNGs (default: frames)
  -threshold float   Scene change detection threshold 0.0–1.0 (default: 0.4)
  -interval float    Temporal sampling interval in seconds (default: 2.0)
  -max-frames int    Maximum output frames; 0 = unlimited (default: 0)
  -hamming int       Hamming distance dedup threshold; 0 = disabled (default: 8)
  -algo string       Hash algorithm: phash or dhash (default: phash)
  -no-scene          Disable scene change detection
  -no-keyframe       Disable keyframe extraction
  -no-temporal       Disable temporal sampling
  -no-mpdecimate     Disable MPDecimate
  -ffmpeg string     Path to ffmpeg binary (default: ffmpeg)
  -ffprobe string    Path to ffprobe binary (default: ffprobe)
  -json              Print frame metadata as JSON to stdout
```

**Examples:**

```bash
# Extract with defaults
extract -i input.mp4 -o ./frames

# Cap to 25 frames, print as JSON
extract -i input.mp4 -o ./frames -max-frames 25 -json

# Scene detection only, lower threshold
extract -i input.mp4 -o ./frames -no-keyframe -no-temporal -no-mpdecimate -threshold 0.25
```

## Configuration reference

| Field | Default | Description |
|---|---|---|
| `Scene` | `true` | Enable scene change detection |
| `Keyframe` | `true` | Enable keyframe (I-frame) extraction |
| `Temporal` | `true` | Enable uniform temporal sampling |
| `MPDecimate` | `true` | Enable MPDecimate |
| `SceneThreshold` | `0.4` | Scene change score threshold [0.0–1.0] |
| `TemporalInterval` | `2.0` | Seconds between temporally sampled frames |
| `MPDecimateHi` | `768` | MPDecimate hi threshold (64×12) |
| `MPDecimateLo` | `320` | MPDecimate lo threshold (64×5) |
| `MPDecimateFrac` | `0.33` | MPDecimate fraction |
| `HashAlgo` | `phash` | Perceptual hash algorithm: `phash` or `dhash` |
| `HammingThreshold` | `8` | Max Hamming distance to consider frames duplicates |
| `HashResizeWidth` | `256` | Width frames are rescaled to before hashing |
| `MaxFrames` | `0` | Output frame cap; 0 = unlimited |
| `WorkDir` | `""` | Output directory; empty = temp dir (auto-cleaned) |
| `Concurrency` | auto | Number of concurrent strategy goroutines |
| `FFmpegPath` | `"ffmpeg"` | Path to ffmpeg binary |
| `FFprobePath` | `"ffprobe"` | Path to ffprobe binary |

## Frame fields

```go
type Frame struct {
    Index        int       // 0-based position in the final ordered set
    TimestampSec float64   // presentation timestamp in seconds
    Strategy     Strategy  // "scene", "keyframe", "temporal", or "mpdecimate"
    Path         string    // absolute path to the extracted PNG
    Hash         uint64    // perceptual hash value
    Video        *VideoInfo
}

type VideoInfo struct {
    Duration  float64 // total duration in seconds
    Width     int
    Height    int
    Codec     string  // e.g. "h264"
    FrameRate float64
    BitRate   int64
}
```

## Error handling

```go
var (
    ErrNoBinaries   // ffmpeg or ffprobe not found in PATH
    ErrNoStrategies // all extraction strategies disabled
    ErrNoFrames     // pipeline produced zero frames
)
```

`FFmpegError` wraps non-zero ffmpeg/ffprobe exits and includes the captured stderr for debugging.

## License

MIT — see [LICENSE](LICENSE).
