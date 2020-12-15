package config

import (
	"github.com/code-ready/crc/pkg/crc/network"
	crcunits "github.com/code-ready/crc/pkg/units"
)

type MachineConfig struct {
	// CRC system bundle
	BundleName string

	// Virtual machine configuration
	Name            string
	Memory          crcunits.Size
	CPUs            int
	DiskSize        crcunits.Size
	ImageSourcePath string
	ImageFormat     string
	SSHKeyPath      string

	// HyperKit specific configuration
	KernelCmdLine string
	Initramfs     string
	Kernel        string

	// Experimental features
	NetworkMode network.Mode
}
