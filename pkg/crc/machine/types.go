package machine

import (
	"github.com/code-ready/crc/pkg/crc/cluster"
	"github.com/code-ready/crc/pkg/crc/network"
	units "github.com/code-ready/crc/pkg/units"
	"github.com/code-ready/machine/libmachine/state"
)

type StartConfig struct {
	// CRC system bundle
	BundlePath string

	// Hypervisor
	Memory   units.Size
	CPUs     int
	DiskSize units.Size

	// Nameserver
	NameServer string

	// User Pull secret
	PullSecret *cluster.PullSecret
}

type ClusterConfig struct {
	ClusterCACert string
	KubeConfig    string
	KubeAdminPass string
	ClusterAPI    string
	WebConsoleURL string
	ProxyConfig   *network.ProxyConfig
}

type StartResult struct {
	Status         state.State
	ClusterConfig  ClusterConfig
	KubeletStarted bool
}

type StopResult struct {
	Name    string
	Success bool
	State   state.State
	Error   string
}

type ClusterStatusResult struct {
	CrcStatus        state.State
	OpenshiftStatus  string
	OpenshiftVersion string
	DiskUse          units.Size
	DiskSize         units.Size
}

type ConsoleResult struct {
	ClusterConfig ClusterConfig
	State         state.State
}
