//go:build darwin || build
// +build darwin build

package vfkit

import (
	"fmt"
)

const (
	VfkitVersion = "0.0.2-dev"
	VfkitCommand = "vfkit"
)

var (
	VfkitDownloadURL = fmt.Sprintf("https://github.com/cfergeau/vfkit/releases/download/v%s/%s", VfkitVersion, VfkitCommand)
)
