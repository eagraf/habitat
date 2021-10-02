package commands

import (
	"github.com/spf13/cobra"
)

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start [process]",
	Short: "Starts a habitat process",
	Long:  `TODO create long description`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		sendRequest("start", args)
	},
}

func init() {
	rootCmd.AddCommand(startCmd)
}
