package cmd

import (
	"errors"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(genereateCmd)
}

var genereateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate a custom bundle from running OpenShift cluster",
	Long:  "Generate a custom bundle from running OpenShift cluster",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runGenerate(args)
	},
}

func runGenerate(arguments []string) error {
	client := newMachine()
	if err := checkIfMachineMissing(client); err != nil {
		return err
	}

	isRunning, err := client.Exists()
	if err != nil {
		return err
	}

	if !isRunning {
		return errors.New("machine is not running")
	}

	err = client.GenerateBundle()
	if err != nil {
		return err
	}
	return nil
}
