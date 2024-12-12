package machine

import (
	"encoding/json"

	macadam "github.com/cfergeau/macadam/pkg/machinedriver"
	"github.com/crc-org/crc/v2/pkg/crc/machine/config"
	crcmac "github.com/crc-org/crc/v2/pkg/crc/machine/macadam"
	"github.com/crc-org/crc/v2/pkg/libmachine"
	"github.com/crc-org/crc/v2/pkg/libmachine/host"
	"github.com/crc-org/machine/libmachine/drivers"
	libdrivers "github.com/crc-org/machine/libmachine/drivers"
	"github.com/pkg/errors"
)

type valueSetter func(driver *libdrivers.VMDriver) bool

func updateDriverValue(host *host.Host, setDriverValue valueSetter) error {
	driver, err := loadDriverConfig(host)
	if err != nil {
		return err
	}
	valueChanged := setDriverValue(driver.VMDriver)
	if !valueChanged {
		return nil
	}

	return updateDriverConfig(host, driver)
}

func setMemory(host *host.Host, memorySize uint) error {
	memorySetter := func(driver *libdrivers.VMDriver) bool {
		if driver.Memory == memorySize {
			return false
		}
		driver.Memory = memorySize
		return true
	}

	return updateDriverValue(host, memorySetter)
}

func setVcpus(host *host.Host, vcpus uint) error {
	vcpuSetter := func(driver *libdrivers.VMDriver) bool {
		if driver.CPU == vcpus {
			return false
		}
		driver.CPU = vcpus
		return true
	}

	return updateDriverValue(host, vcpuSetter)
}

func setDiskSize(host *host.Host, diskSizeGiB uint) error {
	diskSizeSetter := func(driver *libdrivers.VMDriver) bool {
		capacity := config.ConvertGiBToBytes(diskSizeGiB)
		if driver.DiskCapacity == capacity {
			return false
		}
		driver.DiskCapacity = capacity
		return true
	}

	return updateDriverValue(host, diskSizeSetter)
}

func setSharedDirPassword(host *host.Host, password string) error {
	driver, err := loadDriverConfig(host)
	if err != nil {
		return err
	}

	if len(driver.SharedDirs) == 0 {
		return nil
	}

	for i := range driver.SharedDirs {
		driver.SharedDirs[i].Password = password
	}
	return updateDriverStruct(host, driver)
}

func newHost(api libmachine.API, machineConfig config.MachineConfig) (*host.Host, error) {
	json, err := json.Marshal(crcmac.CreateHost(machineConfig))
	if err != nil {
		return nil, errors.New("Failed to marshal driver options")
	}
	return api.NewHost("macadam", "", json)
}

/* FIXME: host.Host is only known here, and libvirt.Driver is only accessible
 * in libvirt/driver_linux.go
 */
func loadDriverConfig(host *host.Host) (*macadam.Driver, error) {
	var macadamDriver macadam.Driver
	err := json.Unmarshal(host.RawDriver, &macadamDriver)

	return &macadamDriver, err
}

func updateDriverConfig(host *host.Host, driver *macadam.Driver) error {
	driverData, err := json.Marshal(driver)
	if err != nil {
		return err
	}
	return host.UpdateConfig(driverData)
}

/*
func (r *RPCServerDriver) SetConfigRaw(data []byte, _ *struct{}) error {
	return json.Unmarshal(data, &r.ActualDriver)
}
*/

func updateDriverStruct(_ *host.Host, _ *macadam.Driver) error {
	// windows was doing: host.Driver = driver
	return drivers.ErrNotImplemented
}
