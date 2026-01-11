package version

import (
	"fmt"
	"runtime"
)

// Version information
var (
	Version   = "1.0.0"
	GitCommit = "unknown"
	BuildTime = "unknown"
	GoVersion = runtime.Version()
)

// GetVersion returns formatted version string
func GetVersion() string {
	return fmt.Sprintf("%s (commit: %s, built: %s, go: %s)", Version, GitCommit, BuildTime, GoVersion)
}
