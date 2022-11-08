package commands

import (
	"fmt"

	"github.com/eagraf/habitat/structs/ctl"
	"github.com/spf13/cobra"
)

// stopCmd represents the stop command
var stopCmd = &cobra.Command{
	Use:   "stop [process]",
	Short: "Stops a habitat process",
	Long:  `TODO create long description`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			fmt.Println("expects a process id as the only arg")
		}
		req := &ctl.StopRequest{
			ProcID: args[0],
		}
		var res ctl.StopResponse
		postRequest(ctl.CommandStop, req, &res)
	},
}

func init() {
	rootCmd.AddCommand(stopCmd)
}
