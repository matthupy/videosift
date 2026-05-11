package ffmpeg

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os/exec"
)

// FFmpegError wraps a non-zero exit status with the captured stderr output.
type FFmpegError struct {
	Args   []string
	Stderr string
	Cause  error
}

func (e *FFmpegError) Error() string {
	return fmt.Sprintf("ffmpeg %v: %v\nstderr: %s", e.Args, e.Cause, e.Stderr)
}

func (e *FFmpegError) Unwrap() error { return e.Cause }

// Cmd is a thin builder around exec.Cmd for ffmpeg/ffprobe invocations.
type Cmd struct {
	bin  string
	args []string
	// Dir sets the working directory for the subprocess. When set, relative
	// paths in arguments (e.g. output patterns, sidecar filenames) are resolved
	// relative to Dir rather than the calling process's working directory.
	Dir string
}

// New returns a Cmd that will run bin with the given arguments.
func New(bin string, args ...string) *Cmd {
	return &Cmd{bin: bin, args: args}
}

// Run executes the command, capturing stderr. Stdout is discarded.
// Returns *FFmpegError on a non-zero exit code.
func (c *Cmd) Run(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, c.bin, c.args...)
	cmd.Dir = c.Dir
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return &FFmpegError{Args: c.args, Stderr: stderr.String(), Cause: err}
	}
	return nil
}

// RunLines executes the command and calls lineFn for each line written to
// stdout. Stderr is captured and included in any returned error.
func (c *Cmd) RunLines(ctx context.Context, lineFn func(string)) error {
	cmd := exec.CommandContext(ctx, c.bin, c.args...)
	cmd.Dir = c.Dir
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("stdout pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return &FFmpegError{Args: c.args, Stderr: stderr.String(), Cause: err}
	}

	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		lineFn(scanner.Text())
	}

	if err := cmd.Wait(); err != nil {
		return &FFmpegError{Args: c.args, Stderr: stderr.String(), Cause: err}
	}
	return nil
}

// LookPath wraps exec.LookPath, returning a descriptive error when the binary
// is not found.
func LookPath(name string) (string, error) {
	path, err := exec.LookPath(name)
	if err != nil {
		return "", fmt.Errorf("%s not found in PATH: %w", name, err)
	}
	return path, nil
}
