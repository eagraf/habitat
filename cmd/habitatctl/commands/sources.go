package commands

import (
	"encoding/json"
	"fmt"

	"github.com/eagraf/habitat/cmd/sources"
	"github.com/eagraf/habitat/structs/ctl"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

// communityCmd represents the community command
var sourcesCmd = &cobra.Command{
	Use:   "sources",
	Short: "Habitat sources is the data access layer on top of JSON, relational, and key-value data.",
	Long:  `Subcommands: FILL ME OUT`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
	},
}

var sourcesReadCmd = &cobra.Command{
	Use:   "read",
	Short: "read a source file",
	Run: func(cmd *cobra.Command, args []string) {

		// TODO: take in community ID flag & set
		id := cmd.Flags().Lookup("id").Value.String()
		node := cmd.Flags().Lookup("node").Value.String()

		log.Debug().Msgf("[Sources] read request for $id: %s at node: %s\n", id, node)

		sourcereq := sources.SourceRequest{
			ID: id,
		}
		b, err := json.Marshal(sourcereq)
		if err != nil {
			fmt.Println("unable to marshal source request ", sourcereq, " err: ", err.Error())
			return
		}

		req := ctl.DataReadRequest{
			Type:   ctl.SourcesRequest,
			Body:   json.RawMessage(b),
			NodeID: node,
		}

		var res ctl.DataReadResponse
		postRequest(ctl.CommandDataServerRead, req, &res)

		fmt.Print(string(res.Data))
	},
}

var sourcesWriteCmd = &cobra.Command{
	Use:   "write",
	Short: "write a source file",
	Run: func(cmd *cobra.Command, args []string) {

		// TODO: take in community ID flag & set
		id := cmd.Flags().Lookup("id").Value.String()
		data := cmd.Flags().Lookup("data").Value.String()

		log.Debug().Msgf("[Sources] write request for $id: %s, data: %s\n", id, data)

		sourcereq := sources.SourceRequest{
			ID: id,
		}
		b, err := json.Marshal(sourcereq)
		if err != nil {
			fmt.Println("unable to marshal source request ", sourcereq, " err: ", err.Error())
			return
		}

		req := ctl.DataWriteRequest{
			Type: ctl.SourcesRequest,
			Body: json.RawMessage(b),
			Data: []byte(data),
		}

		var res ctl.DataWriteResponse
		postRequest(ctl.CommandDataServerWrite, req, &res)
		fmt.Print("success!")
	},
}

func init() {

	sourcesReadCmd.Flags().String("id", "", "$id of the source being read")
	sourcesReadCmd.Flags().String("node", "", "peer id of node to read data from (default local node)")
	sourcesReadCmd.MarkFlagRequired("id")
	sourcesWriteCmd.Flags().String("id", "", "$id of the source being read")
	sourcesWriteCmd.Flags().StringP("data", "d", "", "data to write to the source")
	sourcesWriteCmd.MarkFlagRequired("id")
	sourcesWriteCmd.MarkFlagRequired("data")

	sourcesCmd.AddCommand(sourcesReadCmd)
	sourcesCmd.AddCommand(sourcesWriteCmd)
	rootCmd.AddCommand(sourcesCmd)
}
