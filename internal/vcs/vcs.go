// Package vcs is for reading version numbers from running binary.
package vcs

import "runtime/debug"

func Version() string {
	bi, ok := debug.ReadBuildInfo()
	if ok {
		return bi.Main.Version
	}

	return ""
}
