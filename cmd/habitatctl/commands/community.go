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
	"encoding/base64"
	"encoding/json"
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
		fmt.Printf("%s is an invalid subcommand for community %s\n", args[1], args[0])
	},
}

var communityCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "create a new community",
	Run: func(cmd *cobra.Command, args []string) {

		name := cmd.Flags().Lookup("name")
		if name == nil {
			fmt.Println("name flag needs to be set")
			return
		}

		ipfs, _ := cmd.Flags().GetBool("ipfs")

		req := &ctl.CommunityCreateRequest{
			CommunityName:     name.Value.String(),
			CreateIPFSCluster: ipfs,
		}
		var res ctl.CommunityCreateResponse
		postRequest(ctl.CommandCommunityCreate, req, &res)

		fmt.Println(res.CommunityID)
		fmt.Println(res.JoinToken)
	},
}

var communityJoinCmd = &cobra.Command{
	Use:   "join",
	Short: "join a community",
	Run: func(cmd *cobra.Command, args []string) {

		req := &ctl.CommunityJoinRequest{}
		token, err := cmd.Flags().GetString("token")
		if err != nil {
			printError(err)
		}
		if token == "" {
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
			req.AcceptingNodeAddr = address.Value.String()
			req.CommunityID = communityID.Value.String()
		} else {
			decoded, err := base64.StdEncoding.DecodeString(token)
			if err != nil {
				printError(err)
			}
			var joinInfo ctl.JoinInfo
			err = json.Unmarshal(decoded, &joinInfo)
			if err != nil {
				printError(err)
			}
			req.AcceptingNodeAddr = joinInfo.Address
			req.CommunityID = joinInfo.CommunityID
		}

		var res ctl.CommunityJoinResponse
		postRequest(ctl.CommandCommunityJoin, req, &res)

		fmt.Println(res.AddMemberToken)
	},
}

var communityAddMemberCmd = &cobra.Command{
	Use:   "add",
	Short: "add a member to the community",
	Run: func(cmd *cobra.Command, args []string) {
		req := &ctl.CommunityAddMemberRequest{}
		token, err := cmd.Flags().GetString("token")
		if err != nil {
			printError(err)
		}
		if token == "" {
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
			req.JoiningNodeAddress = address.Value.String()
			req.CommunityID = communityID.Value.String()
			req.NodeID = nodeID.Value.String()
		} else {
			decoded, err := base64.StdEncoding.DecodeString(token)
			if err != nil {
				printError(err)
			}
			var addInfo ctl.AddMemberInfo
			err = json.Unmarshal(decoded, &addInfo)
			if err != nil {
				printError(err)
			}
			req.JoiningNodeAddress = addInfo.Address
			req.CommunityID = addInfo.CommunityID
			req.NodeID = addInfo.NodeID
		}

		postRequest(ctl.CommandCommunityAddMember, req, &ctl.CommunityAddMemberResponse{})
	},
}

var communityProposeTransitionsCmd = &cobra.Command{
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

		req := &ctl.CommunityProposeRequest{
			CommunityID:     communityID.Value.String(),
			StateTransition: []byte(b64Patch),
		}
		postRequest(ctl.CommandCommunityPropose, req, &ctl.CommunityProposeResponse{})
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

		req := &ctl.CommunityStateRequest{
			CommunityID: communityID.Value.String(),
		}

		var res ctl.CommunityStateResponse
		postRequest(ctl.CommandCommunityState, req, &res)

		fmt.Println(string(res.State))
	},
}

var communityListCmd = &cobra.Command{
	Use:   "ls",
	Short: "list the communities that this node is a part of",
	Run: func(cmd *cobra.Command, args []string) {
		req := &ctl.CommunityListRequest{}
		var res ctl.CommunityListResponse

		postRequest(ctl.CommandCommunityList, req, &res)

		fmt.Printf("node id: %s\n", res.NodeID)
		for _, c := range res.Communities {
			fmt.Println(c)
		}
	},
}

func init() {
	communityCreateCmd.Flags().StringP("address", "a", "", "address that this node can be reached at")
	communityCreateCmd.Flags().StringP("name", "n", "", "name of the community being created")
	communityCreateCmd.Flags().Bool("ipfs", false, "create a new IPFS swarm for the community")

	communityJoinCmd.Flags().StringP("address", "a", "", "address that this node can be reached at")
	communityJoinCmd.Flags().StringP("community", "c", "", "id of community to be joined")
	communityJoinCmd.Flags().StringP("token", "t", "", "token to join the community")

	communityAddMemberCmd.Flags().StringP("address", "a", "", "address that this node can be reached at")
	communityAddMemberCmd.Flags().StringP("community", "c", "", "id of community to be joined")
	communityAddMemberCmd.Flags().StringP("node", "n", "", "node id of node that is being added")
	communityAddMemberCmd.Flags().StringP("token", "t", "", "token to add member to the community")

	communityProposeTransitionsCmd.Flags().StringP("community", "c", "", "id of community to be joined")

	communityStateCmd.Flags().StringP("community", "c", "", "id of community to be joined")

	communityCmd.AddCommand(communityCreateCmd)
	communityCmd.AddCommand(communityJoinCmd)
	communityCmd.AddCommand(communityAddMemberCmd)
	communityCmd.AddCommand(communityProposeTransitionsCmd)
	communityCmd.AddCommand(communityStateCmd)
	communityCmd.AddCommand(communityListCmd)

	rootCmd.AddCommand(communityCmd)
}
