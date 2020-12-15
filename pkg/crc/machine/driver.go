package machine

import (
	"github.com/code-ready/crc/pkg/libmachine/host"
	crcunits "github.com/code-ready/crc/pkg/units"
	libmachine "github.com/code-ready/machine/libmachine/drivers"
	"github.com/docker/go-units"
)

type valueSetter func(driver *libmachine.VMDriver) bool

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

func setMemory(host *host.Host, newSize crcunits.Size) error {
	memorySetter := func(driver *libmachine.VMDriver) bool {
		if driver.Memory < 0 {
			return false
		}
		driverSize := crcunits.New(uint64(driver.Memory), units.MiB)
		if driverSize == newSize {
			return false
		}
		driver.Memory = int(newSize.ConvertTo(units.GiB))
		return true
	}

	return updateDriverValue(host, memorySetter)
}

func setVcpus(host *host.Host, vcpus int) error {
	vcpuSetter := func(driver *libmachine.VMDriver) bool {
		if driver.CPU == vcpus {
			return false
		}
		driver.CPU = vcpus
		return true
	}

	return updateDriverValue(host, vcpuSetter)
}

func setDiskSize(host *host.Host, newSize crcunits.Size) error {
	diskSizeSetter := func(driver *libmachine.VMDriver) bool {
		driverSize := crcunits.New(driver.DiskCapacity, units.GiB)
		if driverSize == newSize {
			return false
		}
		driver.DiskCapacity = newSize.ConvertTo(units.GiB)
		return true
	}

	return updateDriverValue(host, diskSizeSetter)
}
