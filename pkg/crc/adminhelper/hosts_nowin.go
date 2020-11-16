// +build !windows

package adminhelper

import (
	crcos "github.com/code-ready/crc/pkg/os"
)

func execute(args ...string) error {
	_, err := crcos.RunWithDefaultLocale(adminHelperPath, args...)
	return err
}
