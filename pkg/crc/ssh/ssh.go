package ssh

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"

	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/crc/logging"
	crcos "github.com/code-ready/crc/pkg/os"
	"github.com/code-ready/machine/libmachine/drivers"
)

type Runner struct {
	driver        drivers.Driver
	privateSSHKey string
}

func CreateRunner(driver drivers.Driver) *Runner {
	return CreateRunnerWithPrivateKey(driver, constants.GetPrivateKeyPath())
}

func CreateRunnerWithPrivateKey(driver drivers.Driver, privateKey string) *Runner {
	return &Runner{driver: driver, privateSSHKey: privateKey}
}

// Create a host using the driver's config
func (runner *Runner) Run(command string, args ...string) (string, error) {
	return runner.runSSHCommandFromDriver(command, false)
}

func (runner *Runner) RunPrivate(command string, args ...string) (string, error) {
	return runner.runSSHCommandFromDriver(command, true)
}

func (runner *Runner) SetPrivateKeyPath(path string) {
	runner.privateSSHKey = path
}

func (runner *Runner) CopyData(data []byte, destFilename string, mode os.FileMode) error {
	logging.Debugf("Creating %s with permissions 0%o in the CRC VM", destFilename, mode)
	base64Data := base64.StdEncoding.EncodeToString(data)
	command := fmt.Sprintf("sudo install -m 0%o /dev/null %s && cat <<EOF | base64 --decode | sudo tee %s\n%s\nEOF", mode, destFilename, destFilename, base64Data)
	_, err := runner.runSSHCommandFromDriver(command, true)

	return err
}

func (runner *Runner) CopyFile(srcFilename string, destFilename string, mode os.FileMode) error {
	data, err := ioutil.ReadFile(srcFilename)
	if err != nil {
		return err
	}
	return runner.CopyData(data, destFilename, mode)
}

func stringFromReadCloser(stream io.ReadCloser) (string, error) {
	buf := new(bytes.Buffer)
	_, err := buf.ReadFrom(stream)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

func (runner *Runner) runSSHCommandFromDriver(command string, runPrivate bool) (string, error) {
	client, err := drivers.GetSSHClientFromDriver(runner.driver, runner.privateSSHKey)
	if err != nil {
		return "", err
	}

	if runPrivate {
		logging.Debugf("About to run SSH command with hidden output")
	} else {
		logging.Debugf("About to run SSH command:\n%s", command)
	}

	stdout, stderr, err := client.Start(command)
	var (
		stdoutStr string
		stderrStr string
	)
	if runPrivate {
		if err != nil {
			logging.Debugf("SSH command failed")
		} else {
			logging.Debugf("SSH command succeeded")
		}
	} else {
		stdoutStr, _ = stringFromReadCloser(stdout)
		stderrStr, _ = stringFromReadCloser(stderr)
		logging.Debugf("SSH command results: err: %v, stdout: %s stderr: %s", err, stdoutStr, stderrStr)
	}

	if err != nil {
		return "", crcos.ExecError{
			Err:    fmt.Errorf("ssh command error: %s - %v", command, err),
			Stdout: stdoutStr,
		}
	}
	defer stdout.Close()
	defer stderr.Close()

	return stdoutStr, nil
}

type remoteCommandRunner struct {
	sshRunner *Runner
}

func (cmdRunner *remoteCommandRunner) Run(cmd string, args ...string) (string, error) {
	commandline := fmt.Sprintf("%s %s", cmd, strings.Join(args, " "))

	return cmdRunner.sshRunner.Run(commandline)
}

func (cmdRunner *remoteCommandRunner) RunPrivate(cmd string, args ...string) (string, error) {
	commandline := fmt.Sprintf("%s %s", cmd, strings.Join(args, " "))

	return cmdRunner.sshRunner.RunPrivate(commandline)
}

func (cmdRunner *remoteCommandRunner) RunPrivileged(reason string, cmdAndArgs ...string) (string, error) {
	commandline := fmt.Sprintf("sudo %s", strings.Join(cmdAndArgs, " "))

	return cmdRunner.sshRunner.Run(commandline)
}

func NewRemoteCommandRunner(runner *Runner) crcos.CommandRunner {
	return &remoteCommandRunner{
		sshRunner: runner,
	}
}
