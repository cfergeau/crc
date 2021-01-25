package adminhelper

import (
	"strings"

	"github.com/code-ready/crc/pkg/os/windows/powershell"
)

func executePrivileged(args ...string) error {
	_, _, err := powershell.ExecuteAsAdmin("modifying hosts file", strings.Join(append([]string{adminHelperPath}, args...), " "))
	return err
}

func execute(args ...string) (string, error) {
	stdout, _, err := powershell.Execute(strings.Join(append([]string{adminHelperPath}, args...), " "))
	return stdout, err
}
