package commands

import (
	"fmt"
	"strings"

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

var commName string

/*var ipfsCmd = &cobra.Command{
	Use:   "ipfs --comm=community_name -- -other -flags ENV_VAR=other_env_vars",
	Short: "Starts a habitat process",
	Long:  `TODO create long description`,
	Args:  cobra.MinimumNArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		client, err := client.NewClient(habitatServiceAddr())
		if err != nil {
			fmt.Println("Error: couldn't connect to habitat service")
			return
		}

		flags, nonflags := parseFlags(cmd.Flags().Args())

		client.WriteRequest(&ctl.Request{
			Command: "start",
			Args:    []string{"ipfs", commName},
			Flags:   flags,
			Env:     append(nonflags),
		})

		res, err := client.ReadResponse()
		if err != nil {
			fmt.Println("Error: couldn't read response from habitat service")
		}
		fmt.Println(res)
	},
}*/

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

		flags, nonflags := parseFlags(args[1:])

		reqArgs := args
		if commName != "" {
			args = append(reqArgs, commName)
		}

		req := &ctl.StartRequest{
			App:   args[0],
			Args:  reqArgs,
			Flags: flags,
			Env:   nonflags,
		}

		var res ctl.StartResponse
		postRequest(ctl.CommandStart, req, &res)

		fmt.Println(res.ProcID)
	},
}

func init() {
	startCmd.PersistentFlags().StringVarP(&commName, "comm", "c", "", "name of community for which to start process")
	//startCmd.AddCommand(ipfsCmd)
	rootCmd.AddCommand(startCmd)
}
