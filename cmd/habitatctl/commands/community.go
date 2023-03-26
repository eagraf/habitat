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
	"errors"
	"fmt"

	client "github.com/eagraf/habitat/pkg/habitat_client"
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
		fmt.Println(cmd.Usage())
	},
}

var communityCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "create a new community",
	Run: func(cmd *cobra.Command, args []string) {

		userIdentity, err := loadUserIdentity(cmd)
		if err != nil {
			printError(fmt.Errorf("error loading user identity: %s", err))
			return
		}

		name := cmd.Flags().Lookup("name")
		if name == nil {
			printError(errors.New("name flag needs to be set"))
		}

		ipfs, _ := cmd.Flags().GetBool("ipfs")

		conn, err := getWebsocketConn(ctl.CommandCommunityCreate)
		if err != nil {
			printError(fmt.Errorf("error establishing websocket connection: %s", err))
		}
		defer conn.Close()

		err = client.WebsocketKeySigningExchange(conn, userIdentity)
		if err != nil {
			printError(fmt.Errorf("error signing new node's certificate: %s", err))
		}

		req := &ctl.CommunityCreateRequest{
			CommunityName:     name.Value.String(),
			CreateIPFSCluster: ipfs,
		}

		err = conn.WriteJSON(req)
		if err != nil {
			printError(err)
		}

		var res ctl.CommunityCreateResponse
		err = conn.ReadJSON(&res)
		if err != nil {
			printError(err)
		}
		if werr := res.GetError(); werr != nil {
			printError(werr)
		}

		pretty, err := json.MarshalIndent(res, "", "    ")
		if err != nil {
			printError(err)
		}
		fmt.Println(string(pretty))
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

		userIdentity, err := loadUserIdentity(cmd)
		if err != nil {
			printError(fmt.Errorf("error loading user identity: %s", err))
			return
		}

		conn, err := getWebsocketConn(ctl.CommandCommunityJoin)
		if err != nil {
			printError(fmt.Errorf("error establishing websocket connection: %s", err))
		}
		defer conn.Close()

		err = client.WebsocketKeySigningExchange(conn, userIdentity)
		if err != nil {
			printError(fmt.Errorf("error signing new node's certificate: %s", err))
		}

		err = conn.WriteJSON(req)
		if err != nil {
			printError(err)
		}

		var res ctl.CommunityJoinResponse
		err = conn.ReadJSON(&res)
		if err != nil {
			printError(err)
		}
		if werr := res.GetError(); werr != nil {
			printError(werr)
		}
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

var communityPSCmd = &cobra.Command{
	Use:   "ps",
	Short: "list the processes that are actively running in this community",
	Run: func(cmd *cobra.Command, args []string) {
		communityID := cmd.Flags().Lookup("community")
		if communityID == nil {
			printError(fmt.Errorf("community flag needs to be set"))
			return
		}

		req := &ctl.CommunityPSRequest{
			CommunityID: communityID.Value.String(),
		}

		var res ctl.CommunityPSResponse
		postRequest(ctl.CommandCommunityPS, req, &res)

		pretty, err := json.MarshalIndent(&res, "", "    ")
		if err != nil {
			printError(fmt.Errorf("error prettifying JSON response: %s", err))
		}
		fmt.Println(string(pretty))
	},
}

var communityStartProcessCmd = &cobra.Command{
	Use:   "start -c <community_id> -n <node 1> -n <node 2> -- <app_name> [args]",
	Short: "start process instances on specific nodes in the community",
	Run: func(cmd *cobra.Command, args []string) {
		communityID := cmd.Flags().Lookup("community")
		if communityID == nil {
			printError(fmt.Errorf("community flag needs to be set"))
			return
		}

		instanceNodes, err := cmd.Flags().GetStringSlice("node")
		if err != nil {
			printError(err)
			return
		}

		dashPosition := cmd.ArgsLenAtDash()
		if dashPosition >= len(cmd.Flags().Args()) {
			printError(fmt.Errorf("Usage: %s", cmd.Use))
			return
		}
		procArgs := cmd.Flags().Args()[cmd.ArgsLenAtDash():]
		env, resArgs := []string{}, []string{}
		if len(procArgs) > 1 {
			env, resArgs = parseEnv(procArgs[1:])
		}

		if len(procArgs) < 1 {
			printError(fmt.Errorf("no app specified"))
			return
		}

		req := &ctl.CommunityStartProcessRequest{
			CommunityID: communityID.Value.String(),
			App:         procArgs[0],
			Args:        resArgs,
			Env:         env,

			InstancesNodes: instanceNodes,
		}

		var res ctl.CommunityStartProcessRequest
		postRequest(ctl.CommandCommunityStartProcess, req, &res)
	},
}

var communityStopProcessCmd = &cobra.Command{
	Use:   "stop -c <community_id> [-n <node_id>...] <process_id>",
	Short: "stop a process on a community. if nodes are specified, only their specific process instances will be stopped",
	Run: func(cmd *cobra.Command, args []string) {
		communityID := cmd.Flags().Lookup("community")
		if communityID == nil {
			printError(fmt.Errorf("community flag needs to be set"))
			return
		}

		instanceNodes, err := cmd.Flags().GetStringSlice("node")
		if err != nil {
			printError(err)
			return
		}

		if len(args) < 1 {
			printError(fmt.Errorf("must supply a process ID to stop"))
		}

		req := &ctl.CommunityStopProcessRequest{
			CommunityID:    communityID.Value.String(),
			ProcessID:      args[0],
			InstancesNodes: instanceNodes,
		}

		var res ctl.CommunityStopProcessResponse
		postRequest(ctl.CommandCommunityStopProcess, req, &res)
	},
}

func init() {
	communityCreateCmd.Flags().StringP("address", "a", "", "address that this node can be reached at")
	communityCreateCmd.Flags().StringP("name", "n", "", "name of the community being created")
	communityCreateCmd.Flags().Bool("ipfs", false, "create a new IPFS swarm for the community")
	addUserFlags(communityCreateCmd)

	communityJoinCmd.Flags().StringP("address", "a", "", "address that this node can be reached at")
	communityJoinCmd.Flags().StringP("community", "c", "", "id of community to be joined")
	communityJoinCmd.Flags().StringP("token", "t", "", "token to join the community")
	addUserFlags(communityJoinCmd)

	communityProposeTransitionsCmd.Flags().StringP("community", "c", "", "id of community to be joined")
	addUserFlags(communityProposeTransitionsCmd)

	communityStateCmd.Flags().StringP("community", "c", "", "id of community to be joined")
	addUserFlags(communityStateCmd)

	communityPSCmd.Flags().StringP("community", "c", "", "id of community to be joined")

	communityStartProcessCmd.Flags().StringP("community", "c", "", "id of community to be joined")
	communityStartProcessCmd.Flags().StringSliceP("node", "n", []string{}, "node ID to have a process instance started on")

	communityStopProcessCmd.Flags().StringP("community", "c", "", "id of community to be joined")
	communityStopProcessCmd.Flags().StringSliceP("node", "n", []string{}, "node ID to have a process instance started on")

	communityCmd.AddCommand(communityCreateCmd)
	communityCmd.AddCommand(communityJoinCmd)
	communityCmd.AddCommand(communityProposeTransitionsCmd)
	communityCmd.AddCommand(communityStateCmd)
	communityCmd.AddCommand(communityListCmd)
	communityCmd.AddCommand(communityPSCmd)
	communityCmd.AddCommand(communityStartProcessCmd)
	communityCmd.AddCommand(communityStopProcessCmd)

	rootCmd.AddCommand(communityCmd)
}
