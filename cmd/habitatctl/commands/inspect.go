package commands

import (
	"encoding/json"
	"fmt"

	"github.com/eagraf/habitat/structs/ctl"
	"github.com/spf13/cobra"
)

var inspectCmd = &cobra.Command{
	Use:   "inspect",
	Short: "inspects a Habitat instance",
	Run: func(cmd *cobra.Command, args []string) {
		var res ctl.InspectResponse
		postRequest(ctl.CommandInspect, &ctl.InspectRequest{}, &res)

		pretty, err := json.MarshalIndent(&res, "", "    ")
		if err != nil {
			printError(err)
		}
		fmt.Println(string(pretty))
	},
}

func init() {
	rootCmd.AddCommand(inspectCmd)
}
