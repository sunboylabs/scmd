// Package version provides version information for scmd
package version

import (
	"fmt"
	"runtime"
)

// Set by ldflags at build time
var (
	Version = "0.5.1"
	Commit  = "none"
	Date    = "unknown"
)

// Info returns full version information
func Info() string {
	return fmt.Sprintf("scmd %s (%s) built %s with %s",
		Version, shortCommit(), Date, runtime.Version())
}

// Short returns just the version
func Short() string {
	return Version
}

// Full returns version with commit
func Full() string {
	return fmt.Sprintf("%s-%s", Version, shortCommit())
}

func shortCommit() string {
	if len(Commit) >= 7 {
		return Commit[:7]
	}
	return Commit
}
