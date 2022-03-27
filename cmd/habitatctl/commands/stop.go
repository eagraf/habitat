package commands

import (
	client "github.com/eagraf/habitat/pkg/habitat_client"
	"github.com/spf13/cobra"
)

// stopCmd represents the stop command
var stopCmd = &cobra.Command{
	Use:   "stop [process]",
	Short: "Stops a habitat process",
	Long:  `TODO create long description`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client.SendRequest("stop", args)
	},
}

func init() {
	rootCmd.AddCommand(stopCmd)
}
