package adminhelper

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/code-ready/crc/pkg/crc/constants"
)

var (
	adminHelperPath = filepath.Join(constants.CrcBinDir, constants.AdminHelperExecutableName)
)

// UpdateHostsFile updates the host's /etc/hosts file with Instance IP.
func UpdateHostsFile(instanceIP string, hostnames ...string) error {
	if err := RemoveFromHostsFile(hostnames...); err != nil {
		return err
	}
	if err := AddToHostsFile(instanceIP, hostnames...); err != nil {
		return err
	}
	return nil
}

func AddToHostsFile(instanceIP string, hostnames ...string) error {
	if err := validateHostnames(hostnames); err != nil {
		return err
	}
	var filtered []string
	for _, hostname := range hostnames {
		stdout, err := execute("contains", instanceIP, hostname)
		if err != nil {
			return err
		}
		if strings.TrimSpace(stdout) == "false" {
			filtered = append(filtered, hostname)
		}
	}
	if len(filtered) == 0 {
		return nil
	}
	return executePrivileged(append([]string{"add", instanceIP}, filtered...)...)
}

func RemoveFromHostsFile(hostnames ...string) error {
	if err := validateHostnames(hostnames); err != nil {
		return err
	}
	return executePrivileged(append([]string{"rm"}, hostnames...)...)
}

func CleanHostsFile() error {
	return executePrivileged([]string{"clean", constants.ClusterDomain, constants.AppsDomain}...)
}

func validateHostnames(hostnames []string) error {
	for _, hostname := range hostnames {
		if !strings.HasSuffix(hostname, constants.ClusterDomain) && !strings.HasSuffix(hostname, constants.AppsDomain) {
			return fmt.Errorf("adding %s to hosts file is denied (outside of %s and %s subdomains)", hostname, constants.ClusterDomain, constants.AppsDomain)
		}
	}
	return nil
}
