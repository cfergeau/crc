package config

import (
	"github.com/crc-org/machine/libmachine/drivers"
)

func ConvertGiBToBytes(gib uint) uint64 {
	return uint64(gib) * 1024 * 1024 * 1024
}

func InitVMDriverFromMachineConfig(machineConfig MachineConfig, driver *drivers.VMDriver) {
	driver.CPU = machineConfig.CPUs
	driver.Memory = machineConfig.Memory
	driver.DiskCapacity = ConvertGiBToBytes(machineConfig.DiskSize)
	driver.BundleName = machineConfig.BundleName
	driver.ImageSourcePath = machineConfig.ImageSourcePath
	driver.ImageFormat = machineConfig.ImageFormat
	sshConfig := drivers.SSHConfig{
		IdentityPath:   machineConfig.SSHKeyPath,
		Port:           22,
		RemoteUsername: "core",
	}
	driver.SSHConfig = &sshConfig
}
