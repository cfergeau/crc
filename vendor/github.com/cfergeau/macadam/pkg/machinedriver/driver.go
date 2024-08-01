/*
Copyright 2021, Red Hat, Inc - All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package macadam

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/containers/common/pkg/strongunits"
	"github.com/containers/podman/v5/pkg/machine"
	"github.com/containers/podman/v5/pkg/machine/define"
	"github.com/containers/podman/v5/pkg/machine/env"
	provider2 "github.com/containers/podman/v5/pkg/machine/provider"
	"github.com/containers/podman/v5/pkg/machine/shim"
	"github.com/containers/podman/v5/pkg/machine/vmconfigs"
	crcos "github.com/crc-org/crc/v2/pkg/os"
	"github.com/crc-org/machine/libmachine/drivers"
	"github.com/crc-org/machine/libmachine/state"
)

const (
	DriverName    = "macadam"
	DriverVersion = "0.0.1"

	DefaultMemory  = 8192
	DefaultCPUs    = 4
	DefaultSSHUser = "core"
)

const (
	// from "github.com/crc-org/crc/v2/pkg/crc/constants"
	DaemonVsockPort = 1024
)

type Driver struct {
	*drivers.VMDriver
	VirtioNet bool

	// TODO: add configuration for this in podman machine
	VsockPath       string
	DaemonVsockPort uint
	QemuGAVsockPort uint

	vmConfig   *vmconfigs.MachineConfig
	vmProvider vmconfigs.VMProvider
}

func NewDriver(hostName, storePath string) *Driver {
	// checks that macdriver.Driver implements the libmachine.Driver interface
	var _ drivers.Driver = &Driver{}

	provider, err := provider2.Get()
	if err != nil {
		return nil
	}
	return &Driver{
		VMDriver: &drivers.VMDriver{
			BaseDriver: &drivers.BaseDriver{
				MachineName: hostName,
				StorePath:   storePath,
			},
			CPU:    DefaultCPUs,
			Memory: DefaultMemory,
		},
		// needed when loading a VM which was created before
		// DaemonVsockPort was introduced
		DaemonVsockPort: DaemonVsockPort,

		vmProvider: provider,
	}
}

// DriverName returns the name of the driver
func (d *Driver) DriverName() string {
	return DriverName
}

// Get Version information
func (d *Driver) DriverVersion() string {
	return DriverVersion
}

// GetIP returns an IP or hostname that this host is available at
// inherited from  libmachine.BaseDriver
// func (d *Driver) GetIP() (string, error)

// GetMachineName returns the name of the machine
// inherited from  libmachine.BaseDriver
// func (d *Driver) GetMachineName() string

// GetBundleName() Returns the name of the unpacked bundle which was used to create this machine
// inherited from  libmachine.BaseDriver
// func (d *Driver) GetBundleName() (string, error)

// PreCreateCheck allows for pre-create operations to make sure a driver is ready for creation
func (d *Driver) PreCreateCheck() error {
	return nil
}

func (d *Driver) getDiskPath() string {
	return d.ResolveStorePath(fmt.Sprintf("%s.img", d.MachineName))
}

func (d *Driver) initOpts() *define.InitOptions {
	initOpts := define.InitOptions{}
	// defaults from cmd/podman/machine/init.go
	initOpts.Name = d.MachineName

	initOpts.CPUS = uint64(d.VMDriver.CPU)
	initOpts.DiskSize = uint64(strongunits.ToGiB(strongunits.B(d.VMDriver.DiskCapacity)))
	initOpts.Memory = uint64(d.VMDriver.Memory)
	initOpts.TimeZone = ""
	initOpts.ReExec = false
	/*
		initOpts.Username = defaultConfig.Machine.User
		initOpts.Image = defaultConfig.Machine.Image
		initOpts.Volumes = defaultConfig.Machine.Volumes.Get()
	*/
	initOpts.Username = "core"
	initOpts.Image = d.getDiskPath()
	initOpts.Volumes = []string{}
	initOpts.USBs = []string{}
	initOpts.IgnitionPath = ""
	initOpts.Rootful = false
	userModeNetworking := false
	initOpts.UserModeNetworking = &userModeNetworking
	// user-mode networking

	return &initOpts
}

func (d *Driver) Reload() error {
	if d.vmProvider == nil {
		provider, err := provider2.Get()
		if err != nil {
			return err
		}
		d.vmProvider = provider
	}
	vmConfig, _, err := shim.VMExists(d.MachineName, []vmconfigs.VMProvider{d.vmProvider})
	if err != nil {
		return err
	}
	d.vmConfig = vmConfig

	return nil
}

func (d *Driver) Create() error {
	if err := d.PreCreateCheck(); err != nil {
		return err
	}

	switch d.ImageFormat {
	case "qcow2":
		// FIXME: the libvirt machine driver uses qemu-img
		if err := crcos.CopyFile(d.ImageSourcePath, d.getDiskPath()); err != nil {
			return err
		}
	default:
		return fmt.Errorf("%s is an unsupported disk image format", d.ImageFormat)
	}

	// Check if machine already exists
	vmConfig, exists, err := shim.VMExists(d.MachineName, []vmconfigs.VMProvider{d.vmProvider})
	if err != nil {
		return err
	}
	// machine exists, return error
	if exists {
		// overwrite vmConfig if machine already exists?
		d.vmConfig = vmConfig
		return fmt.Errorf("%s: %w", d.MachineName, define.ErrVMAlreadyExists)
	}

	/*
		// check if a system connection already exists
		cons, err := registry.PodmanConfig().ContainersConfDefaultsRO.GetAllConnections()
		if err != nil {
			return err
		}
		for _, con := range cons {
			if con.ReadWrite {
				for _, connection := range []string{initOpts.Name, fmt.Sprintf("%s-root", initOpts.Name)} {
					if con.Name == connection {
						return fmt.Errorf("system connection %q already exists. consider a different machine name or remove the connection with `podman system connection rm`", connection)
					}
				}
			}
		}
	*/

	initOpts := d.initOpts()
	for idx, vol := range initOpts.Volumes {
		initOpts.Volumes[idx] = os.ExpandEnv(vol)
	}

	// TODO need to work this back in
	// if finished, err := vm.Init(initOpts); err != nil || !finished {
	// 	// Finished = true,  err  = nil  -  Success! Log a message with further instructions
	// 	// Finished = false, err  = nil  -  The installation is partially complete and podman should
	// 	//                                  exit gracefully with no error and no success message.
	// 	//                                  Examples:
	// 	//                                  - a user has chosen to perform their own reboot
	// 	//                                  - reexec for limited admin operations, returning to parent
	// 	// Finished = *,     err != nil  -  Exit with an error message
	// 	return err
	// }

	err = shim.Init(*initOpts, d.vmProvider)
	if err != nil {
		return err
	}

	/*
		newMachineEvent(events.Init, events.Event{Name: initOpts.Name})
	*/
	fmt.Println("Machine init complete")

	// most likely not needed as libmachine must already be doing this check somehow
	vmConfig, _, err = shim.VMExists(initOpts.Name, []vmconfigs.VMProvider{d.vmProvider})
	if err != nil {
		return err
	}
	d.vmConfig = vmConfig

	return nil
}

// Start a host
func (d *Driver) Start() error {
	machineName := d.vmConfig.Name
	dirs, err := env.GetMachineDirs(d.vmProvider.VMType())
	if err != nil {
		return err
	}
	/*
		mc, err := vmconfigs.LoadMachineByName(machineName, dirs)
		if err != nil {
			return err
		}
	*/

	fmt.Printf("Starting machine %q\n", machineName)

	startOpts := machine.StartOptions{
		NoInfo: false,
		Quiet:  false,
	}
	slog.Info(fmt.Sprintf("SSH config: %v", d.vmConfig.SSH))

	if err := shim.Start(d.vmConfig, d.vmProvider, dirs, startOpts); err != nil {
		return err
	}
	fmt.Printf("Machine %q started successfully\n", machineName)
	//newMachineEvent(events.Start, events.Event{Name: vmName})
	return nil
	/*
		if err := d.recoverFromUncleanShutdown(); err != nil {
			return err
		}

		bootLoader := config.NewLinuxBootloader(
			d.VmlinuzPath,
			"console=hvc0 "+d.Cmdline,
			d.InitrdPath,
		)

		vm := config.NewVirtualMachine(
			uint(d.CPU),
			uint64(d.Memory),
			bootLoader,
		)

		// console
		logFile := d.ResolveStorePath("vfkit.log")
		dev, err := config.VirtioSerialNew(logFile)
		if err != nil {
			return err
		}
		err = vm.AddDevice(dev)
		if err != nil {
			return err
		}

		// network
		// 52:54:00 is the OUI used by QEMU
		const mac = "52:54:00:70:2b:79"
		if d.VirtioNet {
			dev, err = config.VirtioNetNew(mac)
			if err != nil {
				return err
			}
			err = vm.AddDevice(dev)
			if err != nil {
				return err
			}
		}

		// shared directories
		if d.supportsVirtiofs() {
			for _, sharedDir := range d.SharedDirs {
				// TODO: add support for 'mount.ReadOnly'
				// TODO: check format
				dev, err := config.VirtioFsNew(sharedDir.Source, sharedDir.Tag)
				if err != nil {
					return err
				}
				err = vm.AddDevice(dev)
				if err != nil {
					return err
				}
			}
		}

		// entropy
		dev, err = config.VirtioRngNew()
		if err != nil {
			return err
		}
		err = vm.AddDevice(dev)
		if err != nil {
			return err
		}

		// disk
		diskPath := d.getDiskPath()
		dev, err = config.VirtioBlkNew(diskPath)
		if err != nil {
			return err
		}
		err = vm.AddDevice(dev)
		if err != nil {
			return err
		}

		// virtio-vsock device
		dev, err = config.VirtioVsockNew(d.DaemonVsockPort, d.VsockPath, true)
		if err != nil {
			return err
		}
		err = vm.AddDevice(dev)
		if err != nil {
			return err
		}

		// when loading a VM created by a crc version predating this commit,
		// d.QemuGAVsockPort will be missing from ~/.crc/machines/crc/config.json
		// In such a case, assume the VM will not support time sync
		if d.QemuGAVsockPort != 0 {
			timesync, err := config.TimeSyncNew(d.QemuGAVsockPort)
			if err != nil {
				return err
			}
			err = vm.AddDevice(timesync)
			if err != nil {
				return err
			}
		}

		args, err := vm.ToCmdLine()
		if err != nil {
			return err
		}
		process, err := startVfkit(d.VfkitPath, args)
		if err != nil {
			return err
		}

		_ = os.WriteFile(d.getPidFilePath(), []byte(strconv.Itoa(process.Pid)), 0600)

		if !d.VirtioNet {
			return nil
		}

		getIP := func() error {
			d.IPAddress, err = GetIPAddressByMACAddress(mac)
			if err != nil {
				return &RetriableError{Err: err}
			}
			return nil
		}

		if err := RetryAfter(60, getIP, 2*time.Second); err != nil {
			return fmt.Errorf("IP address never found in dhcp leases file %v", err)
		}
		log.Debugf("IP: %s", d.IPAddress)

		return nil
	*/
}

func (d *Driver) GetSharedDirs() ([]drivers.SharedDir, error) {
	return d.SharedDirs, nil
}

func podmanStatusToCrcState(status define.Status) state.State {
	switch status {
	case define.Running:
		return state.Running
	case define.Stopped:
		return state.Stopped
	case define.Starting:
		return state.Running
	case define.Unknown:
		return state.Error
	}

	// unknown state
	return state.Error
}

// GetState returns the state that the host is in (running, stopped, etc)
func (d *Driver) GetState() (state.State, error) {
	if d.vmConfig == nil {
		return state.Stopped, nil
	}
	status, err := d.vmProvider.State(d.vmConfig, false)
	if err != nil {
		return state.Error, err
	}
	return podmanStatusToCrcState(status), nil
	// piggy back on podman machine state
	//return state.Error, fmt.Errorf("GetState() unimplemented")
}

// Kill stops a host forcefully
func (d *Driver) Kill() error {
	return fmt.Errorf("Kill() unimplemented")
}

// Remove a host
func (d *Driver) Remove() error {
	return fmt.Errorf("Remove() unimplemented")
	/*
		s, err := d.GetState()
		if err != nil || s == state.Error {
			log.Debugf("Error checking machine status: %v, assuming it has been removed already", err)
		}
		if s == state.Running {
			if err := d.Kill(); err != nil {
				return err
			}
		}
		return nil
	*/
}

// UpdateConfigRaw allows to change the state (memory, ...) of an already created machine
func (d *Driver) UpdateConfigRaw(rawConfig []byte) error {
	return fmt.Errorf("UpdateConfigRaw() unimplemented")
	/*
		var newDriver Driver
		err := json.Unmarshal(rawConfig, &newDriver)
		if err != nil {
			return err
		}

		err = d.resize(int64(newDriver.DiskCapacity))
		if err != nil {
			log.Debugf("failed to resize disk image: %v", err)
			return err
		}
		*d = newDriver

		return nil
	*/
}

// Stop a host gracefully
func (d *Driver) Stop() error {
	// podman machine stop
	return fmt.Errorf("Stop() unimplemented")
}

func (d *Driver) SSH() drivers.SSHConfig {
	return drivers.SSHConfig{
		IdentityPath:   d.vmConfig.SSH.IdentityPath,
		Port:           d.vmConfig.SSH.Port,
		RemoteUsername: d.vmConfig.SSH.RemoteUsername,
	}
}
