package ping

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/pbv7/pingheat/internal/parser"
)

// Runner executes ping commands and emits samples.
type Runner struct {
	target   string
	interval time.Duration
	parser   parser.Parser
}

// NewRunner creates a new ping runner.
func NewRunner(target string, interval time.Duration) *Runner {
	return &Runner{
		target:   target,
		interval: interval,
		parser:   parser.New(),
	}
}

// Run starts the ping process and sends samples to the channel.
// It blocks until the context is cancelled.
func (r *Runner) Run(ctx context.Context, samples chan<- Sample) error {
	var cmd *exec.Cmd
	args := r.buildArgs()

	if runtime.GOOS == "windows" {
		// Windows: Use cmd.exe to set code page to 437 (US English)
		// This ensures ping output is in English regardless of system locale
		cmdLine := "chcp 437 >nul & ping " + strings.Join(args, " ")
		cmd = exec.CommandContext(ctx, "cmd.exe", "/C", cmdLine)
	} else {
		// Linux/macOS: Force C locale for English output
		cmd = exec.CommandContext(ctx, "ping", args...)
		cmd.Env = append(os.Environ(), "LC_ALL=C", "LANG=C")
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	if err := cmd.Start(); err != nil {
		// Include the full command in the error message for debugging
		return fmt.Errorf("failed to start ping command 'ping %v': %w", args, err)
	}

	// Read stdout in a goroutine
	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := scanner.Text()
			if sample, ok := r.parser.ParseLine(line); ok {
				select {
				case samples <- sample:
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	// Read stderr (mostly for debugging)
	stderrBuf := make([]byte, 0, 1024)
	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			line := scanner.Text()
			stderrBuf = append(stderrBuf, line...)
			stderrBuf = append(stderrBuf, '\n')

			// Parse stderr too - some systems report timeouts here
			if sample, ok := r.parser.ParseLine(line); ok {
				select {
				case samples <- sample:
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	// Wait for process to exit
	err = cmd.Wait()
	if ctx.Err() != nil {
		// Context was cancelled, not an error
		return nil
	}
	if err != nil {
		// Include stderr output in the error message
		if len(stderrBuf) > 0 {
			return fmt.Errorf("ping command failed: %w (stderr: %s)", err, string(stderrBuf))
		}
		return fmt.Errorf("ping command failed with args %v: %w", args, err)
	}
	return nil
}

// buildArgs builds platform-specific ping arguments.
func (r *Runner) buildArgs() []string {
	intervalSec := r.interval.Seconds()

	switch runtime.GOOS {
	case "darwin":
		// macOS: ping -i interval target
		return []string{"-i", formatFloat(intervalSec), r.target}
	case "windows":
		// Windows: ping -t target (continuous ping)
		// Windows doesn't support custom intervals well, so we use -t for continuous
		return []string{"-t", r.target}
	default:
		// Linux: ping -i interval target
		return []string{"-i", formatFloat(intervalSec), r.target}
	}
}

// formatFloat formats a float with minimal precision.
func formatFloat(f float64) string {
	if f == float64(int(f)) {
		return formatInt(int(f))
	}
	// Simple formatting - up to 2 decimal places
	i := int(f)
	frac := int((f - float64(i)) * 100)
	if frac == 0 {
		return formatInt(i)
	}
	if frac%10 == 0 {
		return formatInt(i) + "." + formatInt(frac/10)
	}
	return formatInt(i) + "." + formatInt(frac)
}

func formatInt(i int) string {
	if i == 0 {
		return "0"
	}
	if i < 0 {
		return "-" + formatInt(-i)
	}
	var buf [20]byte
	pos := len(buf)
	for i > 0 {
		pos--
		buf[pos] = byte('0' + i%10)
		i /= 10
	}
	return string(buf[pos:])
}
