package commands

import (
	"fmt"

	"github.com/eagraf/habitat/cmd/habitatctl/client"
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
		client, err := client.NewClient()
		if err != nil {
			fmt.Println("Error: couldn't connect to habitat service")
			return
		}

		client.WriteRequest(&ctl.Request{
			Command: "stop",
			Args:    args,
		})

		res, err := client.ReadResponse()
		if err != nil {
			fmt.Println("Error: couldn't read response from habitat service")
		}
		fmt.Println(res)
	},
}

func init() {
	rootCmd.AddCommand(stopCmd)
}
