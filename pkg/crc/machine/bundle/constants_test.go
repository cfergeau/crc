package bundle

import (
	"runtime"
	"testing"

	"github.com/code-ready/crc/pkg/crc/constants"

	"github.com/stretchr/testify/assert"
)

func TestBundleVersionConsistency(t *testing.T) {
	for arch, downloads := range bundleLocations {
		for os, download := range downloads {
			for preset, location := range download {
				// GetDefaultBundles() does not return information about bundles for other arches, we ignore these for now
				if runtime.GOARCH == arch {
					makefileBundleName := constants.GetDefaultBundles()[preset][os]
					assert.Containsf(t, location.GetURIString(), makefileBundleName, "%s does not end with expected %s", location.GetURIString(), makefileBundleName)
				}
			}
		}
	}
}
