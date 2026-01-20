package parser

import (
	"regexp"
	"strconv"
	"time"

	"github.com/pbv7/pingheat/internal/types"
)

// Windows parses ping output from Windows systems.
// Example: Reply from 8.8.8.8: bytes=32 time=14ms TTL=118
type Windows struct {
	replyPattern   *regexp.Regexp
	timeoutPattern *regexp.Regexp
	seqCounter     int
}

// NewWindows creates a new Windows parser.
func NewWindows() *Windows {
	return &Windows{
		// Windows format: Reply from x.x.x.x: bytes=32 time=14ms TTL=118
		// Note: Windows may show time<1ms for very fast responses
		replyPattern: regexp.MustCompile(`Reply from.*time[<=]?(\d+)\s*ms`),
		// Matches: Request timed out.
		timeoutPattern: regexp.MustCompile(`(?i)request timed out|destination.*unreachable|transmit failed|general failure`),
		seqCounter:     0,
	}
}

// ParseLine parses a single line of Windows ping output.
func (p *Windows) ParseLine(line string) (types.Sample, bool) {
	// Try to match a successful reply
	if matches := p.replyPattern.FindStringSubmatch(line); matches != nil {
		p.seqCounter++
		ms, _ := strconv.Atoi(matches[1])
		rtt := time.Duration(ms) * time.Millisecond
		return types.Sample{
			Timestamp: time.Now(),
			Sequence:  p.seqCounter,
			RTT:       rtt,
			Timeout:   false,
		}, true
	}

	// Check for timeout patterns
	if p.timeoutPattern.MatchString(line) {
		p.seqCounter++
		return types.Sample{
			Timestamp: time.Now(),
			Sequence:  p.seqCounter,
			RTT:       0,
			Timeout:   true,
		}, true
	}

	return types.Sample{}, false
}
