// +build !windows

package adminhelper

import (
	crcos "github.com/code-ready/crc/pkg/os"
)

func executePrivileged(args ...string) error {
	_, _, err := crcos.RunWithDefaultLocale(adminHelperPath, args...)
	return err
}

func execute(args ...string) (string, error) {
	stdout, _, err := crcos.RunWithDefaultLocale(adminHelperPath, args...)
	return stdout, err
}
