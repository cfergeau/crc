package cmd

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/crc/daemonclient"
	crcErrors "github.com/code-ready/crc/pkg/crc/errors"
	"github.com/code-ready/crc/pkg/crc/input"
	"github.com/code-ready/crc/pkg/crc/logging"
	"github.com/spf13/cobra"
)

var clearCache bool

func init() {
	deleteCmd.Flags().BoolVarP(&clearCache, "clear-cache", "", false,
		fmt.Sprintf("Clear the instance cache at: %s", constants.MachineCacheDir))
	addOutputFormatFlag(deleteCmd)
	addForceFlag(deleteCmd)
	rootCmd.AddCommand(deleteCmd)
}

var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete the instance",
	Long:  "Delete the instance",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runDelete(os.Stdout, daemonclient.New(), clearCache, constants.MachineCacheDir, outputFormat != jsonFormat, globalForce, outputFormat)
	},
}

func deleteMachine(daemonClient *daemonclient.Client, clearCache bool, cacheDir string, interactive, force bool) (bool, error) {
	if clearCache {
		if !interactive && !force {
			return false, errors.New("non-interactive deletion requires --force")
		}
		yes := input.PromptUserForYesOrNo("Do you want to delete the instance cache", force)
		if yes {
			_ = os.RemoveAll(cacheDir)
		}
	}

	/*
		if err := checkIfMachineMissing(client); err != nil {
			return false, err
		}
	*/

	if !interactive && !force {
		return false, errors.New("non-interactive deletion requires --force")
	}

	yes := input.PromptUserForYesOrNo("Do you want to delete the instance",
		force)
	if yes {
		defer logging.BackupLogFile()
		return true, daemonClient.APIClient.Delete()
	}
	return false, nil
}

func runDelete(writer io.Writer, daemonClient *daemonclient.Client, clearCache bool, cacheDir string, interactive, force bool, outputFormat string) error {
	machineDeleted, err := deleteMachine(daemonClient, clearCache, cacheDir, interactive, force)
	return render(&deleteResult{
		Success:        err == nil,
		Error:          crcErrors.ToSerializableError(err),
		machineDeleted: machineDeleted,
	}, writer, outputFormat)
}

type deleteResult struct {
	Success        bool                         `json:"success"`
	Error          *crcErrors.SerializableError `json:"error,omitempty"`
	machineDeleted bool
}

func (s *deleteResult) prettyPrintTo(writer io.Writer) error {
	if s.Error != nil {
		return s.Error
	}
	if s.machineDeleted {
		if _, err := fmt.Fprintln(writer, "Deleted the instance"); err != nil {
			return err
		}
	}
	return nil
}
