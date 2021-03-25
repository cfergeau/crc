package cmd

import (
	"github.com/code-ready/crc/pkg/crc/daemon"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(daemonCmd)
}

var daemonCmd = &cobra.Command{
	Use:    "daemon",
	Short:  "Run the crc daemon",
	Long:   "Run the crc daemon",
	Hidden: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		daemon, err := daemon.New(config)
		if err != nil {
			return err
		}
		daemon.SetDebug(isDebugLog())

		if err := daemon.Start(); err != nil {
			return err
		}

		return daemon.Run()
	},
}
