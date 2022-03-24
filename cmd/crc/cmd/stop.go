package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/code-ready/crc/pkg/crc/daemonclient"
	crcErrors "github.com/code-ready/crc/pkg/crc/errors"
	"github.com/code-ready/crc/pkg/crc/input"
	"github.com/spf13/cobra"
)

func init() {
	addOutputFormatFlag(stopCmd)
	addForceFlag(stopCmd)
	rootCmd.AddCommand(stopCmd)
}

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the instance",
	Long:  "Stop the instance",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runStop(os.Stdout, daemonclient.New(), outputFormat != jsonFormat, globalForce, outputFormat)
	},
}

func stopMachine(daemonClient *daemonclient.Client, interactive, force bool) (bool, error) {
	/*
		if err := checkIfMachineMissing(client); err != nil {
			return false, err
		}
	*/
	err := daemonClient.APIClient.Stop()
	if err != nil {
		if !interactive && !force {
			return false, err
		}
		isRunning, statusErr := isRunning(daemonClient)
		// Here we are checking the VM state and if it is still running then
		// Ask user to forcefully power off it.
		if statusErr != nil || isRunning {
			// Most of the time force kill don't work and libvirt throw
			// Device or resource busy error. To make sure we give some
			// graceful time to cluster before kill it.
			yes := input.PromptUserForYesOrNo("Do you want to force power off", force)
			if yes {
				err := daemonClient.APIClient.PowerOff()
				return true, err
			}
		}
		return false, err
	}
	return false, nil
}

func runStop(writer io.Writer, daemonClient *daemonclient.Client, interactive, force bool, outputFormat string) error {
	forced, err := stopMachine(daemonClient, interactive, force)
	return render(&stopResult{
		Success: err == nil,
		Forced:  forced,
		Error:   crcErrors.ToSerializableError(err),
	}, writer, outputFormat)
}

type stopResult struct {
	Success bool                         `json:"success"`
	Forced  bool                         `json:"forced"`
	Error   *crcErrors.SerializableError `json:"error,omitempty"`
}

func (s *stopResult) prettyPrintTo(writer io.Writer) error {
	if s.Error != nil {
		return s.Error
	}
	if s.Forced {
		_, err := fmt.Fprintln(writer, "Forcibly stopped the instance")
		return err
	}
	_, err := fmt.Fprintln(writer, "Stopped the instance")
	return err
}
