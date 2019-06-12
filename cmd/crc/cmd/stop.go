package cmd

import (
	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/crc/errors"
	"github.com/code-ready/crc/pkg/crc/input"
	"github.com/code-ready/crc/pkg/crc/machine"
	"github.com/code-ready/crc/pkg/crc/output"
	"github.com/code-ready/machine/libmachine/state"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(stopCmd)
}

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop cluster",
	Long:  "Stop cluster",
	Run: func(cmd *cobra.Command, args []string) {
		runStop(args)
	},
}

func runStop(arguments []string) {
	const machineName string = constants.DefaultName

	vmState, err := machine.Stop(machineName, isDebugLog())
	if err != nil {
		// Here we are checking the VM state and if it is still running then
		// Ask user to forcefully power off it.
		if vmState == state.Running {
			// Most of the time force kill don't work and libvirt throw
			// Device or resource busy error. To make sure we give some
			// graceful time to cluster before kill it.
			yes := input.PromptUserForYesOrNo("Do you want to force power off", globalForce)
			if yes {
				killVM(machineName)
				errors.Exit(0)
			}
		}
		errors.Exit(1)
	}
	output.Out("CodeReady Containers instance stopped")
}

func killVM(machineName string) {
	err := machine.PowerOff(machineName)
	if err != nil {
		errors.Exit(1)
	}
	output.Out("CodeReady Containers instance forcibly stopped")
}
