package cmd

import (
	"fmt"
	"os"

	"github.com/eagraf/habitat/pkg/dex"
	"github.com/eagraf/habitat/pkg/ipfs"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "dex-ipfs",
	Short: "A DEX driver for IPFS",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Hello!")
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringP("api-port", "a", "5001", "IPFS api port number")
	rootCmd.PersistentFlags().StringP("socket-path", "s", "/tmp/dex-ipfs.sock", "UNIX socket path")

	rootCmd.AddCommand(serverCmd)

	clientCmd := &cobra.Command{
		Use:   "client",
		Short: "DEX IPFS client",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Usage()
		},
	}

	clientCmdFactory := dex.NewClientCmdFactory(getWriteClient)
	clientCmd.AddCommand(clientCmdFactory.SchemaCmd())
	clientCmd.AddCommand(clientCmdFactory.InterfaceCmd())
	clientCmd.AddCommand(clientCmdFactory.ImplementationsCmd())

	rootCmd.AddCommand(clientCmd)
}

func getWriteClient(cmd *cobra.Command) (dex.WriteClient, error) {
	ipfsPort, err := cmd.Flags().GetString("api-port")
	if err != nil {
		log.Fatal().Err(err)
	}
	ipfsClient, err := ipfs.NewClient(fmt.Sprintf("http://localhost:%s/api/v0", ipfsPort))
	if err != nil {
		log.Fatal().Err(err)
	}
	writeClient := &DexIpfsWriteClient{
		ipfsClient: ipfsClient,
	}
	return writeClient, nil
}
