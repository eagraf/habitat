package commands

import (
	"fmt"
	"strings"

	"github.com/eagraf/habitat/structs/ctl"
	"github.com/spf13/cobra"
)

func parseEnv(args []string) ([]string, []string) {
	env := []string{}
	resArgs := []string{}
	for _, arg := range args {
		if strings.Index(arg, "=") >= 1 {
			env = append(env, arg)
		} else {
			resArgs = append(resArgs, arg)
		}
	}
	return env, resArgs
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

		env, resArgs := parseEnv(args[1:])

		reqArgs := args
		if commName != "" {
			args = append(reqArgs, commName)
		}

		req := &ctl.StartRequest{
			App:  args[0],
			Args: resArgs,
			Env:  env,
		}

		var res ctl.StartResponse
		postRequest(ctl.CommandStart, req, &res)

		fmt.Println(res.ProcessInstanceID)
	},
}

func init() {
	startCmd.PersistentFlags().StringVarP(&commName, "comm", "c", "", "name of community for which to start process")
	//startCmd.AddCommand(ipfsCmd)
	rootCmd.AddCommand(startCmd)
}
