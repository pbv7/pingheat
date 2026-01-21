package ping

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"reflect"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/pbv7/pingheat/internal/parser"
)

func TestNormalizeTarget(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		output string
	}{
		{name: "bracketed", input: "[2001:db8::1]", output: "2001:db8::1"},
		{name: "plain", input: "2001:db8::1", output: "2001:db8::1"},
		{name: "empty", input: "", output: ""},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := normalizeTarget(tc.input)
			if got != tc.output {
				t.Fatalf("normalizeTarget(%q) = %q, want %q", tc.input, got, tc.output)
			}
		})
	}
}

func TestIsIPv6Literal(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{name: "ipv6", input: "2001:db8::1", want: true},
		{name: "ipv6-bracketed", input: "[2001:db8::1]", want: true},
		{name: "ipv6-link-local-zone", input: "fe80::1%en0", want: true},
		{name: "ipv4", input: "192.0.2.1", want: false},
		{name: "hostname", input: "example.com", want: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := isIPv6Literal(tc.input)
			if got != tc.want {
				t.Fatalf("isIPv6Literal(%q) = %v, want %v", tc.input, got, tc.want)
			}
		})
	}
}

func TestBuildCommandForOS(t *testing.T) {
	interval := time.Second
	tests := []struct {
		name     string
		goos     string
		target   string
		wantCmd  string
		wantArgs []string
	}{
		{
			name:     "darwin-ipv6",
			goos:     "darwin",
			target:   "2001:db8::1",
			wantCmd:  "ping6",
			wantArgs: []string{"-i", "1", "2001:db8::1"},
		},
		{
			name:     "darwin-ipv4",
			goos:     "darwin",
			target:   "192.0.2.1",
			wantCmd:  "ping",
			wantArgs: []string{"-i", "1", "192.0.2.1"},
		},
		{
			name:     "linux-ipv6",
			goos:     "linux",
			target:   "2001:db8::1",
			wantCmd:  "ping",
			wantArgs: []string{"-6", "-i", "1", "2001:db8::1"},
		},
		{
			name:     "linux-ipv4",
			goos:     "linux",
			target:   "192.0.2.1",
			wantCmd:  "ping",
			wantArgs: []string{"-i", "1", "192.0.2.1"},
		},
		{
			name:     "windows",
			goos:     "windows",
			target:   "example.com",
			wantCmd:  "ping",
			wantArgs: []string{"-t", "example.com"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cmd, args := buildCommandForOS(tc.goos, tc.target, interval)
			if cmd != tc.wantCmd {
				t.Fatalf("buildCommandForOS cmd = %q, want %q", cmd, tc.wantCmd)
			}
			if !reflect.DeepEqual(args, tc.wantArgs) {
				t.Fatalf("buildCommandForOS args = %#v, want %#v", args, tc.wantArgs)
			}
		})
	}
}

func TestValidateWindowsTarget(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		wantOK bool
	}{
		{name: "ipv4", input: "192.0.2.1", wantOK: true},
		{name: "ipv6", input: "2001:db8::1", wantOK: true},
		{name: "ipv6-zone", input: "fe80::1%eth0", wantOK: true},
		{name: "hostname", input: "example.com", wantOK: true},
		{name: "empty", input: "", wantOK: false},
		{name: "cmd-injection", input: "8.8.8.8 & whoami", wantOK: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := validateWindowsTarget(tc.input)
			if (err == nil) != tc.wantOK {
				t.Fatalf("validateWindowsTarget(%q) err=%v, wantOK=%v", tc.input, err, tc.wantOK)
			}
		})
	}
}

func TestRunnerRunParsesOutput(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("helper output uses unix-like ping format")
	}

	stdoutLines := []string{
		"64 bytes from 8.8.8.8: icmp_seq=1 ttl=118 time=14.3 ms",
		"Request timeout for icmp_seq 2",
	}
	stdout := strings.Join(stdoutLines, "\n")

	r := &Runner{
		target:     "example.com",
		interval:   time.Second,
		parser:     parser.New(),
		cmdFactory: testCommandFactory(stdout, "", 0),
	}

	samples := make(chan Sample, 2)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := r.Run(ctx, samples); err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	var got []Sample
	for len(got) < 2 {
		select {
		case s := <-samples:
			got = append(got, s)
		case <-time.After(2 * time.Second):
			t.Fatalf("timed out waiting for samples; got %d", len(got))
		}
	}

	var hasTimeout bool
	var hasSuccess bool
	for _, sample := range got {
		if sample.Timeout {
			hasTimeout = true
		} else {
			hasSuccess = true
		}
	}

	if !hasTimeout || !hasSuccess {
		t.Fatalf("expected both timeout and success samples, got: %+v", got)
	}
}

func testCommandFactory(stdout, stderr string, exitCode int) commandFactory {
	return func(ctx context.Context, name string, args ...string) *exec.Cmd {
		cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=TestHelperProcess", "--")
		cmd.Env = append(os.Environ(),
			"GO_WANT_HELPER_PROCESS=1",
			"PINGHELPER_STDOUT="+stdout,
			"PINGHELPER_STDERR="+stderr,
			fmt.Sprintf("PINGHELPER_EXIT=%d", exitCode),
		)
		return cmd
	}
}

func TestHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}

	if stdout := os.Getenv("PINGHELPER_STDOUT"); stdout != "" {
		_, _ = fmt.Fprint(os.Stdout, stdout)
	}
	if stderr := os.Getenv("PINGHELPER_STDERR"); stderr != "" {
		_, _ = fmt.Fprint(os.Stderr, stderr)
	}

	exitCode := 0
	if raw := os.Getenv("PINGHELPER_EXIT"); raw != "" {
		_, _ = fmt.Sscanf(raw, "%d", &exitCode)
	}
	os.Exit(exitCode)
}
