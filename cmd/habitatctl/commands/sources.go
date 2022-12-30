package commands

import (
	"encoding/json"
	"fmt"

	dataproxy "github.com/eagraf/habitat/cmd/habitat/data_proxy"
	"github.com/eagraf/habitat/cmd/sources"
	"github.com/eagraf/habitat/structs/ctl"
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

		fmt.Printf("Source Read Request for %s\n", id)

		sourcereq := sources.SourceRequest{
			ID: id,
		}
		b, err := json.Marshal(sourcereq)
		if err != nil {
			fmt.Println("unable to marshal source request ", sourcereq, " err: ", err.Error())
			return
		}

		req := dataproxy.ReadRequest{
			Type: dataproxy.SourcesRequest,
			Body: json.RawMessage(b),
		}

		var res dataproxy.ReadResponse
		postRequest(ctl.CommandDataServerRead, req, &res)

		fmt.Printf("Read Data: %s\n", res)
	},
}

var sourcesWriteCmd = &cobra.Command{
	Use:   "write",
	Short: "write a source file",
	Run: func(cmd *cobra.Command, args []string) {

		// TODO: take in community ID flag & set
		id := cmd.Flags().Lookup("id").Value.String()
		data := cmd.Flags().Lookup("data").Value.String()

		fmt.Printf("Source Write Request for %s, %s\n", id, data)

		sourcereq := sources.SourceRequest{
			ID: id,
		}
		b, err := json.Marshal(sourcereq)
		if err != nil {
			fmt.Println("unable to marshal source request ", sourcereq, " err: ", err.Error())
			return
		}

		req := dataproxy.WriteRequest{
			Type: dataproxy.SourcesRequest,
			Body: json.RawMessage(b),
			Data: []byte(data),
		}

		var res dataproxy.WriteResponse
		postRequest(ctl.CommandDataServerWrite, req, &res)
		fmt.Printf("Wrote Data: %s\n", res)
	},
}

func init() {

	sourcesReadCmd.Flags().String("id", "", "id (name) of the source being read")
	sourcesReadCmd.MarkFlagRequired("id")
	sourcesWriteCmd.Flags().String("id", "", "id (name) of the source being read")
	sourcesWriteCmd.Flags().StringP("data", "d", "", "data to write to the source")
	sourcesWriteCmd.MarkFlagRequired("id")
	sourcesWriteCmd.MarkFlagRequired("data")

	sourcesCmd.AddCommand(sourcesReadCmd)
	sourcesCmd.AddCommand(sourcesWriteCmd)
	rootCmd.AddCommand(sourcesCmd)
}
