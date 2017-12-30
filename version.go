package observatory

import (
	"fmt"
	"runtime"
)

var (
	Version = "0.0.1"
	Build   = "DEVELOPMENT"
)

// APIVersion is the current Observatory REST API version.
const APIVersion = 0

// VersionInfo returns human-readable version information.
func VersionInfo() string {
	return fmt.Sprintf("Observatory Coordinator v%s (%s)\n%s %s/%s\n", Version, Build, runtime.Version(), runtime.GOOS, runtime.GOARCH)
}
