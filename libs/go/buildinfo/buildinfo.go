// Package buildinfo exposes metadata injected into distributed executables.
package buildinfo

import "runtime"

// These values are set by scripts/package-apimoon.sh using Go linker flags.
var (
	Version = "dev"
	Commit  = "unknown"
	Date    = "unknown"
)

type Info struct {
	Version string `json:"version"`
	Commit  string `json:"commit"`
	Date    string `json:"build_date"`
	OS      string `json:"os"`
	Arch    string `json:"arch"`
}

func Current() Info {
	return Info{Version: Version, Commit: Commit, Date: Date, OS: runtime.GOOS, Arch: runtime.GOARCH}
}
