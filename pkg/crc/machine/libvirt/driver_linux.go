package libvirt

import (
	macadam "github.com/cfergeau/macadam/pkg/machinedriver"
	"github.com/crc-org/crc/v2/pkg/crc/constants"
	"github.com/crc-org/crc/v2/pkg/crc/machine/config"
)

func CreateHost(machineConfig config.MachineConfig) *macadam.Driver {
	macadamDriver := macadam.NewDriver(machineConfig.Name, constants.MachineBaseDir)

	config.InitVMDriverFromMachineConfig(machineConfig, macadamDriver.VMDriver)

	/*
		if machineConfig.NetworkMode == network.UserNetworkingMode {
			macadamDriver.Network = "" // don't need to attach a network interface
			macadamDriver.VSock = true
		} else {
			macadamDriver.Network = DefaultNetwork
		}
	*/

	//macadamDriver.SharedDirs = configureShareDirs(machineConfig)

	return macadamDriver
}

/*
func CreateHost(machineConfig config.MachineConfig) *libvirt.Driver {
	libvirtDriver := libvirt.NewDriver(machineConfig.Name, constants.MachineBaseDir)

	config.InitVMDriverFromMachineConfig(machineConfig, libvirtDriver.VMDriver)

	if machineConfig.NetworkMode == network.UserNetworkingMode {
		libvirtDriver.Network = "" // don't need to attach a network interface
		libvirtDriver.VSock = true
	} else {
		libvirtDriver.Network = DefaultNetwork
	}

	libvirtDriver.StoragePool = DefaultStoragePool
	libvirtDriver.SharedDirs = configureShareDirs(machineConfig)

	return libvirtDriver
}

func configureShareDirs(machineConfig config.MachineConfig) []drivers.SharedDir {
	var sharedDirs []drivers.SharedDir
	for i, dir := range machineConfig.SharedDirs {
		sharedDir := drivers.SharedDir{
			Source: dir,
			Target: dir,
			Tag:    fmt.Sprintf("dir%d", i),
			Type:   "virtiofs",
		}
		sharedDirs = append(sharedDirs, sharedDir)
	}
	return sharedDirs
}
*/
