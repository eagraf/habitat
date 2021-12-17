package commands

import (
	"fmt"
	"strings"

	"github.com/eagraf/habitat/cmd/habitatctl/client"
	"github.com/eagraf/habitat/structs/ctl"
	"github.com/spf13/cobra"
)

func parseFlags(args []string) ([]string, []string) {
	nonflags := []string{}
	flags := []string{}
	for _, arg := range args {
		if strings.HasPrefix(arg, "--") || strings.HasPrefix(arg, "-") {
			flags = append(flags, arg)
		} else {
			nonflags = append(nonflags, arg)
		}
	}
	return flags, nonflags
}

var procName string

var ipfsCmd = &cobra.Command{
	Use:   "ipfs --name=network_name -- DATA_DIR=/path/to/data IPFS_PATH=/path/to/ipfs/data",
	Short: "Starts a habitat process",
	Long:  `TODO create long description`,
	Args:  cobra.MinimumNArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		client, err := client.NewClient()
		if err != nil {
			fmt.Println("Error: couldn't connect to habitat service")
			return
		}

		flags, nonflags := parseFlags(cmd.Flags().Args())

		client.WriteRequest(&ctl.Request{
			Command: "start",
			Args:    []string{"ipfs", procName},
			Flags:   flags,
			Env:     nonflags,
		})

		res, err := client.ReadResponse()
		if err != nil {
			fmt.Println("Error: couldn't read response from habitat service")
		}
		fmt.Println(res)
	},
}

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start [process] [process args]",
	Short: "Starts a habitat process",
	Long:  `TODO create long description`,
	Args:  cobra.MinimumNArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		if cmd.ArgsLenAtDash() != 1 && cmd.ArgsLenAtDash() != -1 {
			fmt.Println("Error: only one process should be specified before -- number specified: ", cmd.ArgsLenAtDash())
			return
		}
		client, err := client.NewClient()
		if err != nil {
			fmt.Println("Error: couldn't connect to habitat service")
			return
		}

		flags, nonflags := parseFlags(args[1:])

		client.WriteRequest(&ctl.Request{
			Command: "start",
			Args:    args,
			Flags:   flags,
			Env:     nonflags,
		})

		res, err := client.ReadResponse()
		if err != nil {
			fmt.Println("Error: couldn't read response from habitat service")
		}
		fmt.Println(res)
	},
}

func init() {
	ipfsCmd.Flags().StringVarP(&procName, "name", "n", "", "name of process to start")
	ipfsCmd.MarkFlagRequired("name")
	startCmd.AddCommand(ipfsCmd)
	rootCmd.AddCommand(startCmd)

}
