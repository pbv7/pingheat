package parser

import (
	"regexp"
	"strconv"
	"time"

	"github.com/pbv7/pingheat/internal/types"
)

// Linux parses ping output from Linux systems.
// Example: 64 bytes from 8.8.8.8: icmp_seq=1 ttl=118 time=14.3 ms
type Linux struct {
	replyPattern   *regexp.Regexp
	timeoutPattern *regexp.Regexp
}

// NewLinux creates a new Linux parser.
func NewLinux() *Linux {
	return &Linux{
		// Matches: 64 bytes from 8.8.8.8: icmp_seq=1 ttl=118 time=14.3 ms
		replyPattern: regexp.MustCompile(`icmp_seq=(\d+).*time=([0-9.]+)\s*ms`),
		// Matches timeout messages
		timeoutPattern: regexp.MustCompile(`(?i)request timeout|no answer|time.*exceeded|unreachable`),
	}
}

// ParseLine parses a single line of Linux ping output.
func (p *Linux) ParseLine(line string) (types.Sample, bool) {
	// Try to match a successful reply
	if matches := p.replyPattern.FindStringSubmatch(line); matches != nil {
		seq, _ := strconv.Atoi(matches[1])
		rtt, err := parseDuration(matches[2])
		if err != nil {
			return types.Sample{}, false
		}
		return types.Sample{
			Timestamp: time.Now(),
			Sequence:  seq,
			RTT:       rtt,
			Timeout:   false,
		}, true
	}

	// Check for timeout patterns
	if p.timeoutPattern.MatchString(line) {
		return types.Sample{
			Timestamp: time.Now(),
			Sequence:  -1,
			RTT:       0,
			Timeout:   true,
		}, true
	}

	return types.Sample{}, false
}
