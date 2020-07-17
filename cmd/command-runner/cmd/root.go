package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/code-ready/crc/pkg/crc/cluster"
	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/crc/errors"
	"github.com/code-ready/crc/pkg/crc/logging"
	crcssh "github.com/code-ready/crc/pkg/crc/ssh"
	"github.com/code-ready/machine/libmachine"
	"github.com/code-ready/machine/libmachine/host"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "crc-command-runner [command]",
	Short: "Helper for integration tests",
	Long:  `crc-command-runner runs a command in the CRC VM as soon as possible during its boot process.`,
	Args:  cobra.MinimumNArgs(1),
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		runPrerun()
	},
	Run: func(cmd *cobra.Command, args []string) {
		runRoot(strings.Join(args, " "))
	},
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		runPostrun()
	},
}

func init() {
	rootCmd.PersistentFlags().StringVar(&logging.LogLevel, "log-level", constants.DefaultLogLevel, "log level (e.g. \"debug | info | warn | error\")")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		logging.Fatal(err)
	}
}

func runPrerun() {
	// Setting up logrus
	logging.InitLogrus(logging.LogLevel, constants.LogFilePath)
}

func waitForLibmachineHost(client *libmachine.Client) (*host.Host, error) {
	var host *host.Host

	libmachineLoad := func() error {
		var err error
		host, err = client.Load(constants.DefaultName)
		if err != nil {
			return &errors.RetriableError{Err: err}
		}
		return nil
	}

	if err := errors.RetryAfter(60, libmachineLoad, 5*time.Second); err != nil {
		return nil, err
	}

	return host, nil
}

func runRoot(sshCmd string) error {
	logging.Debugf("Creating libmachine connection")
	libMachineAPIClient := libmachine.NewClient(constants.MachineBaseDir, constants.MachineCertsDir)
	defer libMachineAPIClient.Close()
	host, err := waitForLibmachineHost(libMachineAPIClient)
	if err != nil {
		return fmt.Errorf("Can't load crc machine: %v", err)
	}

	privateKeyPath := constants.GetPrivateKeyPath()
	sshRunner := crcssh.CreateRunnerWithPrivateKey(host.Driver, privateKeyPath)
	logging.Debugf("Waiting for ssh")
	if err := cluster.WaitForSsh(sshRunner); err != nil {
		return fmt.Errorf("Could not get ssh access to the cluster: %v", err)
	}
	logging.Debugf("Running ssh command: %s", sshCmd)
	if _, err = sshRunner.Run(sshCmd); err != nil {
		return fmt.Errorf("Failed to run ssh command: %v", err)
	}

	return nil
}

func runPostrun() {
	logging.CloseLogging()
}
