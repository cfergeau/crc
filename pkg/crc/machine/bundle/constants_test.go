package bundle

import (
	"fmt"
	"strings"
	"testing"

	"github.com/crc-org/crc/pkg/crc/preset"
	"github.com/crc-org/crc/pkg/crc/version"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBundleVersionURI(t *testing.T) {

	for _, osInfo := range bundleLocations {
		for _, presetInfo := range osInfo {
			for preset, remoteFile := range presetInfo {
				uriPrefix := getURIPrefix(t, preset)
				assert.True(t, strings.HasPrefix(remoteFile.URI, uriPrefix), fmt.Sprintf("URI %s does not match %s version %s", remoteFile.URI, preset, version.GetBundleVersion(preset)))
			}
		}
	}
}

func getURIPrefix(t *testing.T, preset preset.Preset) string {
	bundleVersion := version.GetBundleVersion(preset)
	require.NotEqual(t, bundleVersion, "0.0.0-unset", fmt.Sprintf("%s version is unset (%s), build flags are incorrect", preset, bundleVersion))
	return fmt.Sprintf("https://mirror.openshift.com/pub/openshift-v4/clients/crc/bundles/%s/%s", preset, bundleVersion)
}
