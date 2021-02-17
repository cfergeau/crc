package adminhelper

import (
	"path/filepath"

	"github.com/code-ready/admin-helper/pkg/types"
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
	return instance().Add(&types.AddRequest{
		IP:    instanceIP,
		Hosts: hostnames,
	})
}

func RemoveFromHostsFile(hostnames ...string) error {
	return instance().Remove(&types.RemoveRequest{
		Hosts: hostnames,
	})
}

func CleanHostsFile() error {
	return execute([]string{"clean", constants.ClusterDomain, constants.AppsDomain}...)
}

type helper interface {
	Add(req *types.AddRequest) error
	Remove(req *types.RemoveRequest) error
}
