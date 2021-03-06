package machine

import (
	"context"

	cmdConfig "github.com/code-ready/crc/cmd/crc/cmd/config"
	crcConfig "github.com/code-ready/crc/pkg/crc/config"
	"github.com/code-ready/crc/pkg/crc/machine/types"
	"github.com/code-ready/crc/pkg/crc/network"
	"github.com/code-ready/machine/libmachine/state"
)

type Client interface {
	GetName() string

	Delete() error
	Exists() (bool, error)
	GetConsoleURL() (*types.ConsoleResult, error)
	IP() (string, error)
	PowerOff() error
	Start(ctx context.Context, startConfig types.StartConfig) (*types.StartResult, error)
	Status() (*types.ClusterStatusResult, error)
	Stop() (state.State, error)
	IsRunning() (bool, error)
}

type client struct {
	name   string
	debug  bool
	config crcConfig.Storage
}

func NewClient(name string, debug bool, config crcConfig.Storage) Client {
	return &client{
		name:   name,
		debug:  debug,
		config: config,
	}
}

func (client *client) GetName() string {
	return client.name
}

func (client *client) useVSock() bool {
	return client.networkMode() == network.VSockMode
}

func (client *client) networkMode() network.Mode {
	return network.ParseMode(client.config.Get(cmdConfig.NetworkMode).AsString())
}

func (client *client) monitoringEnabled() bool {
	return client.config.Get(cmdConfig.EnableClusterMonitoring).AsBool()
}
