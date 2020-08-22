// +build !windows

package preflight

import (
	"fmt"
	"os"
	"syscall"

	"github.com/code-ready/crc/pkg/crc/logging"
	crcos "github.com/code-ready/crc/pkg/os"
)

var nonWinPreflightChecks = [...]Check{
	{
		configKeySuffix:  "check-root-user",
		checkDescription: "Checking if running as non-root",
		check:            checkIfRunningAsNormalUser,
		fixDescription:   "crc should be ran as a normal user",
		flags:            NoFix,
	},
}

func checkIfRunningAsNormalUser() error {
	if os.Geteuid() != 0 {
		return nil
	}
	logging.Debug("Ran as root")
	return fmt.Errorf("crc should be ran as a normal user")
}

func setSuid(path string) error {
	logging.Debugf("Making %s suid", path)

	_, _, err := crcos.RunWithPrivilege(fmt.Sprintf("change ownership of %s", path), "chown", "root", path)
	if err != nil {
		return fmt.Errorf("Unable to set ownership of %s to root: %v", path, err)
	}

	/* Can't do this before the chown as the chown will reset the suid bit */
	_, _, err = crcos.RunWithPrivilege(fmt.Sprintf("set suid for %s", path), "chmod", "u+s,g+x", path)
	if err != nil {
		return fmt.Errorf("Unable to set suid bit on %s: %v", path, err)
	}
	return nil
}

func checkSuid(path string) error {
	fi, err := os.Stat(path)
	if err != nil {
		return err
	}
	if fi.Mode()&os.ModeSetuid == 0 {
		return fmt.Errorf("%s does not have the SUID bit set (%s)", path, fi.Mode().String())
	}
	if fi.Sys().(*syscall.Stat_t).Uid != 0 {
		return fmt.Errorf("%s is not owned by root", path)
	}

	return nil
}
