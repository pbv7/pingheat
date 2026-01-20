package parser

import (
	"testing"
	"time"
)

func TestLinuxParser(t *testing.T) {
	p := NewLinux()

	tests := []struct {
		name      string
		line      string
		wantOK    bool
		wantSeq   int
		wantRTT   time.Duration
		wantTO    bool
	}{
		{
			name:    "standard reply",
			line:    "64 bytes from 8.8.8.8: icmp_seq=1 ttl=118 time=14.3 ms",
			wantOK:  true,
			wantSeq: 1,
			wantRTT: 14300 * time.Microsecond,
			wantTO:  false,
		},
		{
			name:    "reply with hostname",
			line:    "64 bytes from dns.google (8.8.8.8): icmp_seq=5 ttl=118 time=10.1 ms",
			wantOK:  true,
			wantSeq: 5,
			wantRTT: 10100 * time.Microsecond,
			wantTO:  false,
		},
		{
			name:   "timeout",
			line:   "Request timeout for icmp_seq 0",
			wantOK: true,
			wantTO: true,
		},
		{
			name:   "header line",
			line:   "PING google.com (142.250.80.46): 56 data bytes",
			wantOK: false,
		},
		{
			name:   "statistics line",
			line:   "5 packets transmitted, 5 packets received, 0.0% packet loss",
			wantOK: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sample, ok := p.ParseLine(tt.line)
			if ok != tt.wantOK {
				t.Errorf("ParseLine() ok = %v, want %v", ok, tt.wantOK)
				return
			}
			if !ok {
				return
			}
			if sample.Timeout != tt.wantTO {
				t.Errorf("Timeout = %v, want %v", sample.Timeout, tt.wantTO)
			}
			if !tt.wantTO {
				if sample.Sequence != tt.wantSeq {
					t.Errorf("Sequence = %d, want %d", sample.Sequence, tt.wantSeq)
				}
				// Allow 1 microsecond tolerance for floating point parsing
				diff := sample.RTT - tt.wantRTT
				if diff < 0 {
					diff = -diff
				}
				if diff > time.Microsecond {
					t.Errorf("RTT = %v, want %v", sample.RTT, tt.wantRTT)
				}
			}
		})
	}
}

func TestDarwinParser(t *testing.T) {
	p := NewDarwin()

	tests := []struct {
		name      string
		line      string
		wantOK    bool
		wantSeq   int
		wantRTT   time.Duration
		wantTO    bool
	}{
		{
			name:    "standard reply",
			line:    "64 bytes from 8.8.8.8: icmp_seq=0 ttl=118 time=14.236 ms",
			wantOK:  true,
			wantSeq: 0,
			wantRTT: 14236 * time.Microsecond,
			wantTO:  false,
		},
		{
			name:    "reply with more decimal places",
			line:    "64 bytes from 1.1.1.1: icmp_seq=10 ttl=57 time=5.789 ms",
			wantOK:  true,
			wantSeq: 10,
			wantRTT: 5789 * time.Microsecond,
			wantTO:  false,
		},
		{
			name:   "timeout",
			line:   "Request timeout for icmp_seq 5",
			wantOK: true,
			wantTO: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sample, ok := p.ParseLine(tt.line)
			if ok != tt.wantOK {
				t.Errorf("ParseLine() ok = %v, want %v", ok, tt.wantOK)
				return
			}
			if !ok {
				return
			}
			if sample.Timeout != tt.wantTO {
				t.Errorf("Timeout = %v, want %v", sample.Timeout, tt.wantTO)
			}
			if !tt.wantTO {
				if sample.Sequence != tt.wantSeq {
					t.Errorf("Sequence = %d, want %d", sample.Sequence, tt.wantSeq)
				}
				// Allow 1 microsecond tolerance for floating point parsing
				diff := sample.RTT - tt.wantRTT
				if diff < 0 {
					diff = -diff
				}
				if diff > time.Microsecond {
					t.Errorf("RTT = %v, want %v", sample.RTT, tt.wantRTT)
				}
			}
		})
	}
}

func TestWindowsParser(t *testing.T) {
	p := NewWindows()

	tests := []struct {
		name    string
		line    string
		wantOK  bool
		wantRTT time.Duration
		wantTO  bool
	}{
		{
			name:    "standard reply",
			line:    "Reply from 8.8.8.8: bytes=32 time=14ms TTL=118",
			wantOK:  true,
			wantRTT: 14 * time.Millisecond,
			wantTO:  false,
		},
		{
			name:    "fast reply",
			line:    "Reply from 192.168.1.1: bytes=32 time<1ms TTL=64",
			wantOK:  true,
			wantRTT: 1 * time.Millisecond,
			wantTO:  false,
		},
		{
			name:   "timeout",
			line:   "Request timed out.",
			wantOK: true,
			wantTO: true,
		},
		{
			name:   "destination unreachable",
			line:   "Destination host unreachable.",
			wantOK: true,
			wantTO: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sample, ok := p.ParseLine(tt.line)
			if ok != tt.wantOK {
				t.Errorf("ParseLine() ok = %v, want %v", ok, tt.wantOK)
				return
			}
			if !ok {
				return
			}
			if sample.Timeout != tt.wantTO {
				t.Errorf("Timeout = %v, want %v", sample.Timeout, tt.wantTO)
			}
			if !tt.wantTO && sample.RTT != tt.wantRTT {
				t.Errorf("RTT = %v, want %v", sample.RTT, tt.wantRTT)
			}
		})
	}
}
