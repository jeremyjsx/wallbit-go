package wallbit

import (
	"runtime/debug"
)

const modulePath = "github.com/jeremyjsx/wallbit-go"

// Version identifies the SDK release advertised in the default User-Agent
// (see [defaultConfig]) and is the single source of truth that operators
// correlate with traffic in Wallbit's access logs.
//
// It is set at build time via linker flags, typically from CI at release
// time:
//
//	go build -ldflags "-X github.com/jeremyjsx/wallbit-go/wallbit.Version=v1.2.3"
var Version = ""

func resolveVersion() string {
	if Version != "" {
		return Version
	}
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return "dev"
	}
	for _, dep := range info.Deps {
		if dep == nil {
			continue
		}
		if dep.Path == modulePath && dep.Version != "" {
			return dep.Version
		}
	}
	// info.Main describes the binary being built. When someone is running
	// tests or `go run` inside this repository it reports "(devel)", which
	// is useless for a User-Agent; only trust it when a real pseudo-version
	// or tag is present (which happens for users who vendor the SDK into
	// their own main module, or for release builds).
	if info.Main.Path == modulePath && info.Main.Version != "" && info.Main.Version != "(devel)" {
		return info.Main.Version
	}
	return "dev"
}
