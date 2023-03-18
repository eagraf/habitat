package cmd

import (
	"fmt"

	"github.com/rs/zerolog/log"

	"github.com/eagraf/habitat/pkg/dex"
	"github.com/spf13/cobra"
)

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Stat the DEX IPFS driver server",
	Run: func(cmd *cobra.Command, args []string) {
		socketPath, err := cmd.Flags().GetString("socket-path")
		if err != nil {
			log.Fatal().Msgf("error getting socket path: %s", err)
		}
		fmt.Println(socketPath)

		ipfsPort, err := cmd.Flags().GetString("api-port")
		if err != nil {
			log.Fatal().Msgf("error getting IPFS port: %s", err)
		}

		driver, err := NewIPFSDexDriver(ipfsPort)
		if err != nil {
			log.Fatal().Msgf("error creating IPFS DEX driver: %s", err)
		}

		server, err := dex.NewServer(socketPath, driver)
		if err != nil {
			log.Fatal().Msgf("error starting DEX server: %s", err)
		}

		server.Start()
	},
}

func init() {
}
