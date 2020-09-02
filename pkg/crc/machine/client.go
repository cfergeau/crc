package machine

import (
	"fmt"

	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/crc/machine/bundle"
	"github.com/code-ready/crc/pkg/crc/machine/config"
	"github.com/code-ready/crc/pkg/crc/ssh"

	"github.com/code-ready/machine/libmachine"
	"github.com/code-ready/machine/libmachine/drivers"
	"github.com/code-ready/machine/libmachine/host"
	"github.com/code-ready/machine/libmachine/state"
)

type Client interface {
	Delete(deleteConfig DeleteConfig) (DeleteResult, error)
	Exists(name string) (bool, error)
	GetConsoleURL(consoleConfig ConsoleConfig) (ConsoleResult, error)
	IP(ipConfig IPConfig) (IPResult, error)
	PowerOff(powerOff PowerOffConfig) (PowerOffResult, error)
	Start(startConfig StartConfig) (StartResult, error)
	Status(statusConfig ClusterStatusConfig) (ClusterStatusResult, error)
	Stop(stopConfig StopConfig) (StopResult, error)
}

type client struct {
	machineName string
	apiClient   *libmachine.Client
	host        *host.Host
}

func NewClient() Client {
	return &client{
		apiClient: libmachine.NewClient(constants.MachineBaseDir, constants.MachineCertsDir),
	}
}

func (client *client) Close() {
	if client.apiClient != nil {
		client.apiClient.Close()
	}
}

func (client *client) Save() error {
	if client.host != nil {
		return fmt.Errorf("Invalid client.host")
	}

	return client.apiClient.Save(client.host)
}

func (client *client) SetMachineName(name string) {
	client.machineName = name
}

func (client *client) getDriver() (drivers.Driver, error) {
	host, err := client.getHost()
	if err != nil {
		return nil, err
	}

	return host.Driver, nil
}

func (client *client) getHost() (*host.Host, error) {
	if client.host != nil {
		return client.host, nil
	}

	host, err := client.apiClient.Load(client.machineName)
	if err == nil {
		client.host = host
	}

	return host, nil
}

func (client *client) createHost(machineConfig config.MachineConfig) error {
	host, err := newHost(client.apiClient, machineConfig)
	if err != nil {
		return fmt.Errorf("Error creating new host: %s", err)
	}
	if err := client.apiClient.Create(host); err != nil {
		return fmt.Errorf("Error creating the VM: %s", err)
	}
	client.host = host

	return nil
}

func (client *client) getBundleMetadata() (*bundle.CrcBundleInfo, error) {
	driver, err := client.getDriver()
	if err != nil {
		return nil, err
	}
	bundleName, err := driver.GetBundleName()
	if err != nil {
		err := fmt.Errorf("Error getting bundle name from CodeReady Containers instance, make sure you ran 'crc setup' and are using the latest bundle")
		return nil, err
	}
	metadata, err := bundle.GetCachedBundleInfo(bundleName)
	if err != nil {
		return nil, err
	}

	return metadata, err
}

func (client *client) Kill() error {
	host, err := client.getHost()
	if err != nil {
		return err
	}
	return host.Kill()
}

func (client *client) Remove() error {
	driver, err := client.getDriver()
	if err != nil {
		return err
	}
	if err := driver.Remove(); err != nil {
		return err
	}

	if err := client.apiClient.Remove(client.machineName); err != nil {
		return err
	}

	return nil
}

func (client *client) GetIP() (string, error) {
	driver, err := client.getDriver()
	if err != nil {
		return "", err
	}
	return driver.GetIP()
}

func (client *client) GetState() (state.State, error) {
	driver, err := client.getDriver()
	if err != nil {
		return state.Error, err
	}
	return driver.GetState()
}

func (client *client) GetSSHRunner() (*ssh.Runner, error) {
	driver, err := client.getDriver()
	if err != nil {
		return nil, err
	}
	return ssh.CreateRunner(driver), nil
}
