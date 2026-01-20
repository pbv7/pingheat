package parser

import (
	"runtime"
	"time"

	"github.com/pbv7/pingheat/internal/types"
)

// Parser parses ping output lines into samples.
type Parser interface {
	// ParseLine parses a single line of ping output.
	// Returns a sample and true if the line contained timing info.
	// Returns zero sample and false for non-timing lines.
	ParseLine(line string) (types.Sample, bool)
}

// New returns a Parser appropriate for the current platform.
func New() Parser {
	switch runtime.GOOS {
	case "darwin":
		return NewDarwin()
	case "windows":
		return NewWindows()
	default: // linux and others
		return NewLinux()
	}
}

// parseDuration parses a floating point milliseconds string into time.Duration.
func parseDuration(ms string) (time.Duration, error) {
	var f float64
	_, err := parseFloat(ms, &f)
	if err != nil {
		return 0, err
	}
	return time.Duration(f * float64(time.Millisecond)), nil
}

// parseFloat parses a string to float64.
func parseFloat(s string, f *float64) (int, error) {
	var n int
	_, err := parseFloatManual(s, f)
	return n, err
}

// parseFloatManual manually parses float to avoid importing strconv everywhere.
func parseFloatManual(s string, f *float64) (int, error) {
	var result float64
	var decimal float64 = 1
	var inDecimal bool
	var negative bool
	i := 0

	if len(s) > 0 && s[0] == '-' {
		negative = true
		i++
	}

	for ; i < len(s); i++ {
		c := s[i]
		if c >= '0' && c <= '9' {
			if inDecimal {
				decimal *= 10
				result += float64(c-'0') / decimal
			} else {
				result = result*10 + float64(c-'0')
			}
		} else if c == '.' && !inDecimal {
			inDecimal = true
		} else {
			break
		}
	}

	if negative {
		result = -result
	}
	*f = result
	return i, nil
}
