package config

import (
	"github.com/code-ready/machine/libmachine/drivers"
)

func ConvertGBToBytes(gigabytes int) uint64 {
	return uint64(gigabytes) * 1000_000_000
}

func InitVMDriverFromMachineConfig(machineConfig MachineConfig, driver *drivers.VMDriver) {
	driver.CPU = machineConfig.CPUs
	driver.Memory = machineConfig.Memory
	driver.DiskCapacity = ConvertGBToBytes(machineConfig.DiskSize)
	driver.BundleName = machineConfig.BundleName
	driver.ImageSourcePath = machineConfig.ImageSourcePath
	driver.ImageFormat = machineConfig.ImageFormat
	driver.SSHKeyPath = machineConfig.SSHKeyPath
}
