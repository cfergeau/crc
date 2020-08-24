package oc

import (
	"path/filepath"

	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/crc/ssh"
	crcos "github.com/code-ready/crc/pkg/os"
)

type Config struct {
	Runner         crcos.CommandRunner
	OcBinaryPath   string
	KubeconfigPath string
	Context        string
	Cluster        string
}

// UseOcWithConfig return the oc binary along with valid kubeconfig
func UseOCWithConfig(machineName string) Config {
	return Config{
		Runner:         crcos.NewLocalCommandRunner(),
		OcBinaryPath:   filepath.Join(constants.CrcOcBinDir, constants.OcBinaryName),
		KubeconfigPath: filepath.Join(constants.MachineInstanceDir, machineName, "kubeconfig"),
		Context:        constants.DefaultContext,
		Cluster:        constants.DefaultName,
	}
}

func (oc Config) runCommand(redactedData []string, args ...string) (string, error) {
	if oc.Context != "" {
		args = append(args, "--context", oc.Context)
	}
	if oc.Cluster != "" {
		args = append(args, "--cluster", oc.Cluster)
	}
	if oc.KubeconfigPath != "" {
		args = append(args, "--kubeconfig", oc.KubeconfigPath)
	}

	return oc.Runner.RunRedacted(redactedData, oc.OcBinaryPath, args...)
}

func (oc Config) RunOcCommand(args ...string) (string, error) {
	return oc.runCommand([]string{}, args...)
}

func (oc Config) RunOcCommandRedacted(redactedData []string, args ...string) (string, error) {
	return oc.runCommand(redactedData, args...)
}

func UseOCWithSSH(sshRunner *ssh.Runner) Config {
	return Config{
		Runner:         ssh.NewRemoteCommandRunner(sshRunner),
		OcBinaryPath:   "oc",
		KubeconfigPath: "/opt/kubeconfig",
		Context:        constants.DefaultContext,
		Cluster:        constants.DefaultName,
	}
}
