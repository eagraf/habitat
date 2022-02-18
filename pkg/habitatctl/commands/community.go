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
				SendRequest(ctl.CommandCommunityAddMember, []string{args[1], args[2]})
				return
			case "propose":
				SendRequest(ctl.CommandCommunityPropose, []string{})
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

		address := cmd.Flags().Lookup("address")
		if address == nil {
			fmt.Println("address flag needs to be set")
			return
		}

		SendRequest(ctl.CommandCommunityCreate, []string{address.Value.String()})
	},
}

var communityJoinCmd = &cobra.Command{
	Use:   "join",
	Short: "join a community",
	Run: func(cmd *cobra.Command, args []string) {

		address := cmd.Flags().Lookup("address")
		if address == nil {
			fmt.Println("address flag needs to be set")
			return
		}

		communityID := cmd.Flags().Lookup("community")
		if communityID == nil {
			fmt.Println("community flag needs to be set")
			return
		}

		SendRequest(ctl.CommandCommunityJoin, []string{address.Value.String(), communityID.Value.String()})
	},
}

var communityAddMemberCmd = &cobra.Command{
	Use:   "add",
	Short: "add a member to the community",
	Run: func(cmd *cobra.Command, args []string) {
		address := cmd.Flags().Lookup("address")
		if address == nil {
			fmt.Println("address flag needs to be set")
			return
		}

		communityID := cmd.Flags().Lookup("community")
		if communityID == nil {
			fmt.Println("community flag needs to be set")
			return
		}

		nodeID := cmd.Flags().Lookup("node")
		if communityID == nil {
			fmt.Println("node flag needs to be set")
			return
		}

		SendRequest(ctl.CommandCommunityAddMember, []string{
			communityID.Value.String(),
			nodeID.Value.String(),
			address.Value.String(),
		})
	},
}

var communityProposeTransitionCmd = &cobra.Command{
	Use:   "propose <json_patch_b64>",
	Short: "propose a transition to this community's state",
	Run: func(cmd *cobra.Command, args []string) {
		communityID := cmd.Flags().Lookup("community")
		if communityID == nil {
			fmt.Println("community flag needs to be set")
			return
		}

		if len(args) < 1 {
			fmt.Println("must supply a base64 encoded JSON patch as the first argument")
			return
		}
		b64Patch := args[0]

		SendRequest(ctl.CommandCommunityPropose, []string{communityID.Value.String(), b64Patch})
	},
}

var communityStateCmd = &cobra.Command{
	Use:   "state",
	Short: "get the state of the community as a JSON object",
	Run: func(cmd *cobra.Command, args []string) {
		communityID := cmd.Flags().Lookup("community")
		if communityID == nil {
			fmt.Println("community flag needs to be set")
			return
		}

		SendRequest(ctl.CommandCommunityState, []string{communityID.Value.String()})
	},
}

var communityListCmd = &cobra.Command{
	Use:   "ls",
	Short: "list the communities that this node is a part of",
	Run: func(cmd *cobra.Command, args []string) {
		SendRequest(ctl.CommandCommunityList, []string{})
	},
}

func init() {
	communityCreateCmd.Flags().StringP("address", "a", "", "address that this node can be reached at")

	communityJoinCmd.Flags().StringP("address", "a", "", "address that this node can be reached at")
	communityJoinCmd.Flags().StringP("community", "c", "", "id of community to be joined")

	communityAddMemberCmd.Flags().StringP("address", "a", "", "address that this node can be reached at")
	communityAddMemberCmd.Flags().StringP("community", "c", "", "id of community to be joined")
	communityAddMemberCmd.Flags().StringP("node", "n", "", "node id of node that is being added")

	communityProposeTransitionCmd.Flags().StringP("community", "c", "", "id of community to be joined")

	communityStateCmd.Flags().StringP("community", "c", "", "id of community to be joined")

	communityCmd.AddCommand(communityCreateCmd)
	communityCmd.AddCommand(communityJoinCmd)
	communityCmd.AddCommand(communityAddMemberCmd)
	communityCmd.AddCommand(communityProposeTransitionCmd)
	communityCmd.AddCommand(communityStateCmd)
	communityCmd.AddCommand(communityListCmd)

	rootCmd.AddCommand(communityCmd)
}
