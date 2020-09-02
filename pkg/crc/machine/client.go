package machine

import (
	"fmt"
	"runtime"

	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/crc/machine/bundle"
	"github.com/code-ready/crc/pkg/crc/machine/config"
	crcssh "github.com/code-ready/crc/pkg/crc/ssh"
	crcos "github.com/code-ready/crc/pkg/os"

	"github.com/code-ready/crc/pkg/crc/logging"

	"github.com/code-ready/machine/libmachine"
	"github.com/code-ready/machine/libmachine/drivers"
	"github.com/code-ready/machine/libmachine/host"
	"github.com/code-ready/machine/libmachine/log"
	"github.com/code-ready/machine/libmachine/ssh"
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
	host, err := client.getHost()
	if err != nil {
		return err
	}

	return client.apiClient.Save(host)
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

func (client *client) GetSSHRunner() (*crcssh.Runner, error) {
	driver, err := client.getDriver()
	if err != nil {
		return nil, err
	}
	return crcssh.CreateRunner(driver), nil
}

func (client *client) GetSSHRunnerWithPrivateKey(privateKeyPath string) (*crcssh.Runner, error) {
	driver, err := client.getDriver()
	if err != nil {
		return nil, err
	}
	return crcssh.CreateRunnerWithPrivateKey(driver, privateKeyPath), nil
}

func (client *client) HostStop() error {
	host, err := client.getHost()
	if err != nil {
		return err
	}

	return host.Stop()
}

func (client *client) DriverStart() error {
	driver, err := client.getDriver()
	if err != nil {
		return err
	}

	return driver.Start()
}

func setMachineLogging(logs bool) error {
	if !logs {
		log.SetDebug(true)
		logfile, err := logging.OpenLogFile(constants.LogFilePath)
		if err != nil {
			return err
		}
		log.SetOutWriter(logfile)
		log.SetErrWriter(logfile)
	} else {
		log.SetDebug(true)
	}
	return nil
}

func unsetMachineLogging() {
	logging.CloseLogFile()
}

func init() {
	// Force using the golang SSH implementation for windows
	if runtime.GOOS == crcos.WINDOWS.String() {
		ssh.SetDefaultClient(ssh.Native)
	}
}

func IsRunning(st state.State) bool {
	return st == state.Running
}
