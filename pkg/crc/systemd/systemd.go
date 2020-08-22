package systemd

import (
	"fmt"

	"github.com/code-ready/crc/pkg/crc/ssh"
	"github.com/code-ready/crc/pkg/crc/systemd/actions"
	"github.com/code-ready/crc/pkg/crc/systemd/states"
	crcos "github.com/code-ready/crc/pkg/os"
)

type Commander struct {
	commandRunner crcos.CommandRunner
}

func NewInstanceSystemdCommander(sshRunner *ssh.Runner) *Commander {
	return &Commander{
		commandRunner: ssh.NewRemoteCommandRunner(sshRunner),
	}
}

func (c Commander) Enable(name string) error {
	_, err := c.service(name, actions.Enable)
	return err
}

func (c Commander) Disable(name string) error {
	_, err := c.service(name, actions.Disable)
	return err
}

func (c Commander) Reload(name string) error {
	_ = c.DaemonReload()
	_, err := c.service(name, actions.Reload)
	return err
}

func (c Commander) Restart(name string) error {
	_ = c.DaemonReload()
	_, err := c.service(name, actions.Restart)
	return err
}

func (c Commander) Start(name string) error {
	_ = c.DaemonReload()
	_, err := c.service(name, actions.Start)
	return err
}

func (c Commander) Stop(name string) error {
	_, err := c.service(name, actions.Stop)
	return err
}

func (c Commander) Status(name string) (states.State, error) {
	return c.service(name, actions.Status)

}

func (c Commander) DaemonReload() error {
	stdOut, err := c.commandRunner.RunPrivileged("executing systemctl daemon-reload command", "systemctl", "daemon-reload")
	if err != nil {
		return fmt.Errorf("Executing systemctl daemon-reload failed: %s: %v", stdOut, err)
	}
	return nil
}

func (c Commander) service(name string, action actions.Action) (states.State, error) {
	var (
		stdOut string
		err    error
	)
	if action.IsPriviledged() {
		msg := fmt.Sprintf("executing systemctl %s %s", action.String(), name)
		stdOut, err = c.commandRunner.RunPrivileged(msg, "systemctl", action.String(), name)
	} else {
		stdOut, err = c.commandRunner.Run("systemctl", action.String(), name)
	}

	if err != nil {
		return states.Error, fmt.Errorf("Executing systemctl action failed: %s: %v", stdOut, err)
	}

	return states.Compare(stdOut), nil
}
