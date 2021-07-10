package commands

import (
	"fmt"

	"github.com/eagraf/habitat/cmd/habitatctl/client"
	"github.com/eagraf/habitat/structs/ctl"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Starts a habitat process",
	Long:  `TODO create long description`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("start called")

		client, err := client.NewClient()
		if err != nil {
			log.Error().Msgf("error creating client: %s", err)
		}

		client.WriteRequest(&ctl.Request{
			Command: "start",
			Text:    "Hola senor",
		})

		res, err := client.ReadResponse()
		if err != nil {
			log.Error().Msgf("error reading response: %s", err)
		}
		fmt.Println(res)
	},
}

func init() {
	rootCmd.AddCommand(startCmd)
}
