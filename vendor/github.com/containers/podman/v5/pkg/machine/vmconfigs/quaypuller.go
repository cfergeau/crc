package vmconfigs

import (
	"fmt"
	"runtime"

	"github.com/containers/podman/v5/pkg/machine/define"
	"github.com/containers/podman/v5/pkg/machine/env"
	"github.com/containers/podman/v5/pkg/machine/shim/diskpull"
)

type QuayPuller struct {
	localPath     *define.VMFile
	sourceURI     string
	vmType        define.VMType
	machineConfig *MachineConfig
	machineDirs   *define.MachineDirs
}

func NewQuayPuller(vmType define.VMType, mc *MachineConfig) (*QuayPuller, error) {
	puller := QuayPuller{
		vmType:        vmType,
		machineConfig: mc,
	}

	dirs, err := env.GetMachineDirs(vmType)
	if err != nil {
		return nil, err
	}
	puller.machineDirs = dirs

	return &puller, nil
}

func (puller QuayPuller) SetSourceURI(uri string) {
	puller.sourceURI = uri
}

func imageExtension(vmType define.VMType) string {
	switch vmType {
	case define.QemuVirt:
		return ".qcow2"
	case define.AppleHvVirt, define.LibKrun:
		return ".raw"
	case define.HyperVVirt:
		return ".vhdx"
	case define.WSLVirt:
		return ""
	default:
		return ""
	}
}

func localImagePath(machineDirs *define.MachineDirs, name string, imageExtension string) (*define.VMFile, error) {
	return machineDirs.DataDir.AppendToNewVMFile(fmt.Sprintf("%s-%s%s", name, runtime.GOARCH, imageExtension), nil)
}

func (puller QuayPuller) LocalPath() (*define.VMFile, error) {
	return localImagePath(puller.machineDirs, puller.machineConfig.Name, imageExtension(puller.vmType))
}

func (puller QuayPuller) Download() error {
	imagePath, err := puller.LocalPath()
	if err != nil {
		return err
	}
	return diskpull.GetDisk(puller.sourceURI, puller.machineDirs, imagePath, puller.vmType, puller.machineConfig.Name)
}
