// +build !linux

package machine

import (
	"github.com/code-ready/crc/pkg/crc/machine/bundle"
)

func copyDiskImage(dirName string, crcBundleMetadata *bundle.CrcBundleInfo) error {
	return nil
}
