package ping

import "testing"

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
