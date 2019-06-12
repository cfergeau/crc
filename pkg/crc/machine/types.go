package machine

import (
	"github.com/code-ready/machine/libmachine/state"
)

type StartConfig struct {
	Name string

	// CRC system bundle
	BundlePath string

	// Hypervisor
	VMDriver string
	Memory   int
	CPUs     int

	// Nameserver
	NameServer string

	// Machine log output
	Debug bool

	// User Pull secret
	PullSecret string
}

type ClusterConfig struct {
	KubeConfig    string
	KubeAdminPass string
	ClusterAPI    string
	WebConsoleURL string
}

type StartResult struct {
	Error          error
	Status         state.State
	ClusterConfig  ClusterConfig
	KubeletStarted bool
}

type StopConfig struct {
	Name  string
	Debug bool
}

type PowerOffConfig struct {
	Name string
}

type DeleteConfig struct {
	Name string
}

type IpConfig struct {
	Name  string
	Debug bool
}

type IpResult struct {
	Name    string
	IP      string
	Success bool
	Error   string
}

type ClusterStatusConfig struct {
	Name string
}

type ClusterStatusResult struct {
	Name            string
	CrcStatus       string
	OpenshiftStatus string
	DiskUse         int64
	DiskSize        int64
	Error           string
	Success         bool
}

type ConsoleConfig struct {
	Name string
}

type ConsoleResult struct {
	URL     string
	Success bool
	Error   string
}
