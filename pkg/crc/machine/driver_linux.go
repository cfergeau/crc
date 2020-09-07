package machine

import (
	"encoding/json"
	"errors"

	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/crc/logging"
	"github.com/code-ready/crc/pkg/crc/machine/config"
	"github.com/code-ready/crc/pkg/crc/machine/libvirt"
	machineLibvirt "github.com/code-ready/machine/drivers/libvirt"
	"github.com/code-ready/machine/libmachine"
	"github.com/code-ready/machine/libmachine/host"
)

func newHost(api libmachine.API, machineConfig config.MachineConfig) (*host.Host, error) {
	json, err := json.Marshal(libvirt.CreateHost(machineConfig))
	if err != nil {
		return nil, errors.New("Failed to marshal driver options")
	}
	return api.NewHost("libvirt", constants.CrcBinDir, json)
}

/* FIXME: host.Host is only known here, and libvirt.Driver is only accessible
 * in libvirt/driver_linux.go
 */
func loadDriverConfig(host *host.Host) (*machineLibvirt.Driver, error) {
	var libvirtDriver machineLibvirt.Driver
	logging.Infof("> loadDriverConfig")
	logging.Infof("RawDriver: %s", host.RawDriver)
	err := json.Unmarshal(host.RawDriver, &libvirtDriver)

	return &libvirtDriver, err
}

func updateDriverConfig(host *host.Host, driver *machineLibvirt.Driver) error {
	logging.Infof("> updateDriverConfig")
	driverData, err := json.Marshal(driver)
	if err != nil {
		logging.Infof("E updateDriverConfig: %v", err)
		return err
	}

	logging.Infof("= updateDriverConfig %s", string(driverData))
	err = host.UpdateConfig(driverData)

	logging.Infof("< updateDriverConfig")
	return err
}

/*
func (r *RPCServerDriver) SetConfigRaw(data []byte, _ *struct{}) error {
	return json.Unmarshal(data, &r.ActualDriver)
}
*/
