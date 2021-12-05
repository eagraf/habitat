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

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start [process] -- [env vars] [flags]",
	Short: "Starts a habitat process",
	Long:  `TODO create long description`,
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		process := args[0]
		fmt.Println("start process ", process)
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
		fmt.Println("flags ", flags)
		fmt.Println("non flags", nonflags)

		client.WriteRequest(&ctl.Request{
			Command: "start",
			Args:    []string{args[0]},
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
	rootCmd.AddCommand(startCmd)
}
