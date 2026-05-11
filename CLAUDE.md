# CLAUDE.md — videosift

## Project overview

`videosift` is a Go library and CLI tool for extracting the minimum set of visually unique frames from a video file. Primary use case: feeding frames into image-based content moderation ML models. Module path: `github.com/matthupy/videosift`.

FFmpeg/FFprobe are invoked via `os/exec` — no CGo. The only third-party dependencies are `goimagehash` (perceptual hashing) and `golang.org/x/sync` (errgroup).

## Build and test

```powershell
go build ./...
go test ./... -v
go vet ./...
```

All commands must be run from the repo root (`C:\Users\Matt\source\repos\videosift`). On a fresh shell on Windows, refresh PATH first:
```powershell
$env:PATH = [System.Environment]::GetEnvironmentVariable("PATH","Machine") + ";" + [System.Environment]::GetEnvironmentVariable("PATH","User")
```

## Architecture

```
Extract(ctx, videoPath, cfg)
  │
  ├─ probe()              ffprobe → VideoInfo
  ├─ runStrategies()      fan-out via errgroup (up to cfg.Concurrency goroutines)
  │   ├─ runScene()       select=gt(scene\,N)
  │   ├─ runKeyframe()    select=eq(pict_type\,I)
  │   ├─ runTemporal()    fps=1/N
  │   └─ runMPDecimate()  mpdecimate=hi:lo:frac
  ├─ merge()              sort by (timestampSec, strategyPrecedence); collapse same-ts duplicates
  ├─ hammingDedup()       O(N²) Hamming distance gate
  ├─ capFrames()          uniform-stride cap to cfg.MaxFrames
  └─ finalizeFrames()     rename surviving PNGs → frame_%05d.png; delete unselected
```

### Key files

| File | Role |
|---|---|
| `extract.go` | `Extract()` entry point and pipeline orchestration |
| `types.go` | All public types: `Config`, `Frame`, `VideoInfo`, `Strategy`, `HashAlgo`, `DefaultConfig()` |
| `errors.go` | `FFmpegError`, sentinel errors |
| `probe.go` | `probe()` — ffprobe JSON parsing → `VideoInfo` |
| `strategies.go` | Four `run*()` functions sharing `runStrategy()` |
| `hash.go` | `computeHash()`, `hammingDistance()` |
| `dedup.go` | `merge()`, `hammingDedup()`, `capFrames()` |
| `internal/ffmpeg/cmd.go` | `Cmd` builder with `Dir` support, `Run()`, `RunLines()` |
| `internal/ffmpeg/parse.go` | `ParseMetadataFile()`, `ParseRationalFPS()` |
| `cmd/extract/main.go` | CLI tool |

### Timestamp capture

Each FFmpeg invocation uses `metadata=print:file=meta.txt` in the filter graph to write a sidecar file alongside the PNG frames. `ParseMetadataFile` reads the sidecar and returns `map[frameIndex]ptsTimeSec`. Frame indices match the `%05d` in the PNG filenames because FFmpeg is invoked with `-start_number 0`.

### Windows path escaping

FFmpeg is run with `cmd.Dir` set to the strategy subdirectory (e.g. `WorkDir/scene/`). Output paths in FFmpeg arguments are relative (`frame_%05d.png`, `meta.txt`). This avoids Windows drive-letter colons and backslashes appearing in FFmpeg filter graph strings, which would require complex escaping.

## Implementation traps

1. **`-vsync 0` is mandatory** for all `select`-based filters; without it FFmpeg duplicates or drops frames unexpectedly.
2. **`select+scene` filter order** — never insert `setpts` before `select=gt(scene,...)`, it resets the score calculation.
3. **`-start_number 0`** — FFmpeg's default output numbering starts at 1; the metadata sidecar frame index starts at 0. Always pass `-start_number 0` or there is an off-by-one.
4. **Strategy subdirectories** — each strategy writes to `WorkDir/<strategy>/`. Never share a directory between concurrent strategies.
5. **Context cancellation** — `exec.CommandContext` kills the process; `runStrategy` removes the strategy subdir on any error to avoid partially-written PNGs reaching `collectFrames`.
6. **`mpdecimate` output volume** — unlike the other strategies, `mpdecimate` output cardinality is unpredictable (0–100% of input frames). `MPDecimateHi/Lo/Frac` are exposed in `Config` for tuning.
7. **Short videos + temporal sampling** — `fps=1/N` on a video shorter than N seconds yields 0 frames. This is not an error; other strategies still contribute frames.
8. **Hash resize** — `HashResizeWidth` (default 256) adds a `scale=256:-1` step to every filter chain. Removing it will make hashing much slower on high-resolution sources.

## Adding a new extraction strategy

1. Add a `StrategyXxx Strategy = "xxx"` constant in `types.go` and assign it a `precedence()` value.
2. Add a `bool` toggle field to `Config`.
3. Write `runXxx(ctx, videoPath, cfg) ([]candidateFrame, error)` in `strategies.go` — call `runStrategy()` with the appropriate filter string.
4. Wire it into `runStrategies()` in `extract.go` and `applyDefaults()` if it needs a default.
5. Add a `-no-xxx` flag to `cmd/extract/main.go`.

## Dependencies

```
github.com/corona10/goimagehash v1.1.0   perceptual hashing (pHash, dHash)
golang.org/x/sync v0.7.0                 errgroup for concurrent strategies
```
