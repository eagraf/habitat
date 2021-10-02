/*
Copyright © 2021 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package commands

import (
	"fmt"

	"github.com/eagraf/habitat/structs/ctl"
	"github.com/spf13/cobra"
)

// communityCmd represents the community command
var communityCmd = &cobra.Command{
	Use:   "community",
	Short: "Habitat communities allow you to run software across multiple nodes",
	Long: `Subcommands:
	create <name>
	join <name>
	<community_id> add  <member_id>
	<community_id> propose <data>
`,
	Run: func(cmd *cobra.Command, args []string) {

		if len(args) < 1 {
			fmt.Println("No subcommand specified")
			return
		}
		switch args[0] {
		default:
			// Assume the community id was specified
			if len(args) < 2 {
				fmt.Printf("No subcommand for community %s specified\n", args[0])
				return
			}
			switch args[1] {
			case "add":
				if len(args) < 3 {
					fmt.Printf("No member id specified for community add command")
				}
				sendRequest(ctl.CommandCommunityAddMember, []string{args[1], args[2]})
				return
			case "propose":
				sendRequest(ctl.CommandCommunityPropose, []string{})
				return
			default:
				fmt.Printf("%s is an invalid subcommand for community %s\n", args[1], args[0])
				return
			}
		}
	},
}

var communityCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "create a new community",
	Run: func(cmd *cobra.Command, args []string) {
		sendRequest(ctl.CommandCommunityCreate, []string{})
	},
}

var communityJoinCmd = &cobra.Command{
	Use:   "join",
	Short: "join a community",
	Run: func(cmd *cobra.Command, args []string) {
		sendRequest(ctl.CommandCommunityJoin, []string{args[0]})
	},
}

func init() {
	communityCmd.AddCommand(communityCreateCmd)
	communityCmd.AddCommand(communityJoinCmd)
	rootCmd.AddCommand(communityCmd)
}
