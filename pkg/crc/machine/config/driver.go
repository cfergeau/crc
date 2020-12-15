package config

import (
	crcunits "github.com/code-ready/crc/pkg/units"

	"github.com/code-ready/machine/libmachine/drivers"
	units "github.com/docker/go-units"
)

func ConvertGiBToBytes(gib int) uint64 {
	return uint64(gib) * 1024 * 1024 * 1024
}

func InitVMDriverFromMachineConfig(machineConfig MachineConfig, driver *drivers.VMDriver) {
	driver.CPU = machineConfig.CPUs
	driver.Memory = int(machineConfig.Memory.ConvertTo(units.MiB))
	driver.DiskCapacity = machineConfig.DiskSize.ConvertTo(crcunits.Bytes)
	driver.BundleName = machineConfig.BundleName
	driver.ImageSourcePath = machineConfig.ImageSourcePath
	driver.ImageFormat = machineConfig.ImageFormat
}
