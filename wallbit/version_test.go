package wallbit

import (
	"strings"
	"testing"
)

func TestResolveVersionLdflagsOverrideWins(t *testing.T) {
	prev := Version
	t.Cleanup(func() { Version = prev })

	Version = "v9.9.9"
	if got := resolveVersion(); got != "v9.9.9" {
		t.Fatalf("resolveVersion() = %q, want %q", got, "v9.9.9")
	}
}

func TestResolveVersionFallsBackToDevWhenUnset(t *testing.T) {
	prev := Version
	t.Cleanup(func() { Version = prev })

	// When Version is empty, resolveVersion consults runtime/debug build
	// info. Inside `go test` running in this repo the main module is
	// modulePath with Main.Version == "(devel)", which the function
	// explicitly rejects as unusable; we therefore expect the "dev"
	// literal. This encodes the contract "never return an empty string",
	// which matters because the value ships in an HTTP header.
	Version = ""
	got := resolveVersion()
	if got == "" {
		t.Fatal("resolveVersion() returned empty string; header would be malformed")
	}
	if got != "dev" {
		t.Fatalf("resolveVersion() = %q in a devel build, want %q", got, "dev")
	}
}

func TestDefaultUserAgentEmbedsResolvedVersion(t *testing.T) {
	prev := Version
	t.Cleanup(func() { Version = prev })

	Version = "v1.2.3"
	ua := defaultUserAgent()
	// Assert both the product token and the version segment so a future
	// refactor that changes the format (e.g. adds a URL suffix) still
	// catches a regression where the version is silently dropped.
	if !strings.HasPrefix(ua, "wallbit-go-sdk/") {
		t.Fatalf("User-Agent %q does not start with product token", ua)
	}
	if !strings.Contains(ua, "v1.2.3") {
		t.Fatalf("User-Agent %q does not contain resolved version", ua)
	}
}
