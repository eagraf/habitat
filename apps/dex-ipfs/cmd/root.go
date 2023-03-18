package cmd

import (
	"fmt"
	"os"

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
}
