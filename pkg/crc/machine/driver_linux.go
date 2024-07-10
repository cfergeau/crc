package machine

import (
	"encoding/json"
	"errors"

	macadam "github.com/cfergeau/macadam/pkg/machinedriver"
	"github.com/crc-org/crc/v2/pkg/crc/machine/config"
	"github.com/crc-org/crc/v2/pkg/crc/machine/libvirt"
	"github.com/crc-org/crc/v2/pkg/libmachine"
	"github.com/crc-org/crc/v2/pkg/libmachine/host"
	"github.com/crc-org/machine/libmachine/drivers"
)

func newHost(api libmachine.API, machineConfig config.MachineConfig) (*host.Host, error) {
	json, err := json.Marshal(libvirt.CreateHost(machineConfig))
	if err != nil {
		return nil, errors.New("Failed to marshal driver options")
	}
	return api.NewHost("macadam", "", json)
}

/* FIXME: host.Host is only known here, and libvirt.Driver is only accessible
 * in libvirt/driver_linux.go
 */
func loadDriverConfig(host *host.Host) (*macadam.Driver, error) {
	var libvirtDriver macadam.Driver
	err := json.Unmarshal(host.RawDriver, &libvirtDriver)

	return &libvirtDriver, err
}

func updateDriverConfig(host *host.Host, driver *macadam.Driver) error {
	driverData, err := json.Marshal(driver)
	if err != nil {
		return err
	}
	return host.UpdateConfig(driverData)
}

func updateKernelArgs(_ *virtualMachine) error {
	return nil
}

/*
func (r *RPCServerDriver) SetConfigRaw(data []byte, _ *struct{}) error {
	return json.Unmarshal(data, &r.ActualDriver)
}
*/

func updateDriverStruct(_ *host.Host, _ *macadam.Driver) error {
	return drivers.ErrNotImplemented
}
