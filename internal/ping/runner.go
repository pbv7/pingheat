package ping

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/pbv7/pingheat/internal/parser"
)

// Runner executes ping commands and emits samples.
type Runner struct {
	target     string
	interval   time.Duration
	parser     parser.Parser
	cmdFactory commandFactory
}

// NewRunner creates a new ping runner.
func NewRunner(target string, interval time.Duration) *Runner {
	return &Runner{
		target:     target,
		interval:   interval,
		parser:     parser.New(),
		cmdFactory: exec.CommandContext,
	}
}

// Run starts the ping process and sends samples to the channel.
// It blocks until the context is cancelled.
func (r *Runner) Run(ctx context.Context, samples chan<- Sample) error {
	var cmd *exec.Cmd
	cmdFactory := r.commandFactory()
	var cmdName string
	var args []string
	target := normalizeTarget(r.target)

	if runtime.GOOS == "windows" {
		// Windows: Use cmd.exe to set code page to 437 (US English)
		// This ensures ping output is in English regardless of system locale
		if err := validateWindowsTarget(target); err != nil {
			return err
		}
		cmdLine := "chcp 437 >nul & ping -t " + quoteCmdArg(target)
		cmdName = "cmd.exe"
		args = []string{"/C", cmdLine}
		cmd = cmdFactory(ctx, cmdName, args...)
	} else {
		// Linux/macOS: Force C locale for English output
		cmdName, args = r.buildCommand(target)
		cmd = cmdFactory(ctx, cmdName, args...)
		if cmd.Env == nil {
			cmd.Env = os.Environ()
		}
		cmd.Env = append(cmd.Env, "LC_ALL=C", "LANG=C")
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
		return fmt.Errorf("failed to start ping command '%s %v': %w", cmdName, args, err)
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
		return fmt.Errorf("ping command failed (%s %v): %w", cmdName, args, err)
	}
	return nil
}

// buildCommand builds platform-specific ping command and arguments.
func (r *Runner) buildCommand(target string) (string, []string) {
	return buildCommandForOS(runtime.GOOS, target, r.interval)
}

func buildCommandForOS(goos, target string, interval time.Duration) (string, []string) {
	intervalSec := interval.Seconds()

	switch goos {
	case "darwin":
		// macOS: ping6 handles IPv6 literals; ping handles IPv4/hostnames.
		if isIPv6Literal(target) {
			return "ping6", []string{"-i", formatFloat(intervalSec), target}
		}
		return "ping", []string{"-i", formatFloat(intervalSec), target}
	case "windows":
		// Windows: ping -t target (continuous ping)
		// Windows doesn't support custom intervals well, so we use -t for continuous
		return "ping", []string{"-t", target}
	default:
		// Linux: ping -i interval target
		args := []string{"-i", formatFloat(intervalSec), target}
		if isIPv6Literal(target) {
			return "ping", append([]string{"-6"}, args...)
		}
		return "ping", args
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

var windowsTargetRe = regexp.MustCompile(`\A[-A-Za-z0-9._:%:]+\z`)

func validateWindowsTarget(target string) error {
	if target == "" {
		return fmt.Errorf("target host required")
	}
	if !windowsTargetRe.MatchString(target) {
		return fmt.Errorf("target contains unsupported characters for Windows ping")
	}
	return nil
}

func quoteCmdArg(arg string) string {
	escaped := strings.ReplaceAll(arg, "%", "^%")
	return `"` + escaped + `"`
}

func isIPv6Literal(target string) bool {
	host := strings.Trim(target, "[]")
	if zone := strings.Index(host, "%"); zone != -1 {
		host = host[:zone]
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.To4() == nil
}

func normalizeTarget(target string) string {
	if len(target) >= 2 && strings.HasPrefix(target, "[") && strings.HasSuffix(target, "]") {
		return target[1 : len(target)-1]
	}
	return target
}

func (r *Runner) commandFactory() commandFactory {
	if r.cmdFactory != nil {
		return r.cmdFactory
	}
	return exec.CommandContext
}

type commandFactory func(ctx context.Context, name string, args ...string) *exec.Cmd
