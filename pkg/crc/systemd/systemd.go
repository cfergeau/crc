package systemd

import (
	"errors"
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
		commandRunner: sshRunner,
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
	_, _, err := c.commandRunner.RunPrivileged("Executing systemctl daemon-reload command", "systemctl", "daemon-reload")
	if err != nil {
		return fmt.Errorf("Executing systemctl daemon-reload failed: %v", err)
	}
	return nil
}

func (c Commander) service(name string, action actions.Action) (states.State, error) {
	var (
		stdOut string
		err    error
	)

	if action.IsPriviledged() {
		msg := fmt.Sprintf("Executing systemctl %s %s", action.String(), name)
		stdOut, _, err = c.commandRunner.RunPrivileged(msg, "systemctl", action.String(), name)
	} else {
		stdOut, _, err = c.commandRunner.Run("systemctl", action.String(), name)
	}

	if err != nil {
		state := states.Compare(stdOut)
		if state != states.Unknown {
			return state, nil
		}
		var execErr *crcos.ExecError
		if errors.As(err, &execErr) {
			state = states.Compare(execErr.Stderr)
			if state == states.NotFound {
				return state, nil
			}
		}

		return states.Error, fmt.Errorf("Executing systemctl action failed: %v", err)
	}

	return states.Compare(stdOut), nil
}

type systemctlUserRunner struct {
	runner crcos.CommandRunner
}

func (userRunner *systemctlUserRunner) Run(command string, args ...string) (string, string, error) {
	if command != "systemctl" {
		return "", "", fmt.Errorf("Invalid command: '%s'", command)
	}
	return userRunner.runner.Run("systemctl", append([]string{"--user"}, args...)...)
}

func (userRunner *systemctlUserRunner) RunPrivate(command string, args ...string) (string, string, error) {
	if command != "systemctl" {
		return "", "", fmt.Errorf("Invalid command: '%s'", command)
	}
	return userRunner.runner.RunPrivate("systemctl", append([]string{"--user"}, args...)...)
}

func (userRunner *systemctlUserRunner) RunPrivileged(reason string, cmdAndArgs ...string) (string, string, error) {
	command := cmdAndArgs[0]
	args := cmdAndArgs[1:]
	if command != "systemctl" {
		return "", "", fmt.Errorf("Invalid command: '%s'", command)
	}
	return userRunner.runner.Run("systemctl", append([]string{"--user"}, args...)...)
}

func (c *Commander) User() *Commander {
	return &Commander{
		commandRunner: &systemctlUserRunner{
			c.commandRunner,
		},
	}
}
