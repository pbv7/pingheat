package version

import "testing"

func TestInfo(t *testing.T) {
	Version = "v1.2.3"
	Commit = "abc123"
	BuildTime = "now"

	got := Info()
	want := "v1.2.3 (abc123) built at now"
	if got != want {
		t.Fatalf("Info() = %q, want %q", got, want)
	}
}
