//go:build !darwin
// +build !darwin

package preflight

import (
	"github.com/crc-org/crc/v2/pkg/crc/constants"
	crcpreset "github.com/crc-org/crc/v2/pkg/crc/preset"
)

func getAllInstanceNamesToDelete() []string {
	instances := []string{}
	for _, preset := range crcpreset.AllPresets() {
		instances = append(instances, constants.InstanceName(preset))
	}
	// add the name of old 'crc' instance before preset based instance names
	return append(instances, "crc")
}
