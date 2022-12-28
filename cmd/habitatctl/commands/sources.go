package commands

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	dataproxy "github.com/eagraf/habitat/cmd/habitat/data_proxy"
	"github.com/eagraf/habitat/cmd/sources"
	"github.com/eagraf/habitat/pkg/compass"
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
		host := cmd.Flags().Lookup("host").Value.String()
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
		b2, err := json.Marshal(req)
		if err != nil {
			fmt.Println("unable to marshal source request ", req, " err: ", err.Error())
			return
		}

		target := fmt.Sprintf("http://%s:%s/read", host, compass.SourcesServerPort())
		rsp, err := http.Post(target, "application/json", bytes.NewBuffer(b2))
		if err != nil {
			log.Error().Msgf("error writing POST to data proxy: %s", err.Error())
			return
		}

		slurp, err := ioutil.ReadAll(rsp.Body)
		if err != nil {
			log.Error().Msgf("error reading response from data proxy: %s", err.Error())
			return
		}

		fmt.Printf("Read Data: %s\n", string(slurp))
	},
}

var sourcesWriteCmd = &cobra.Command{
	Use:   "write",
	Short: "write a source file",
	Run: func(cmd *cobra.Command, args []string) {

		// TODO: take in community ID flag & set
		host := cmd.Flags().Lookup("host").Value.String()
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
		b2, err := json.Marshal(req)
		if err != nil {
			fmt.Println("unable to marshal source request ", req, " err: ", err.Error())
			return
		}

		target := fmt.Sprintf("http://%s:%s/write", host, compass.SourcesServerPort())
		rsp, err := http.Post(target, "application/json", bytes.NewBuffer(b2))
		if err != nil {
			log.Error().Msgf("error writing POST to data proxy: %s", err.Error())
			return
		}

		slurp, err := ioutil.ReadAll(rsp.Body)
		if err != nil {
			log.Error().Msgf("error reading response from data proxy: %s", err.Error())
			return
		}

		fmt.Printf("Wrote Data: %s\n", string(slurp))
	},
}

func init() {

	sourcesReadCmd.Flags().String("id", "", "id (name) of the source being read")
	sourcesReadCmd.MarkFlagRequired("id")
	sourcesWriteCmd.Flags().String("id", "", "id (name) of the source being read")
	sourcesWriteCmd.Flags().StringP("data", "d", "", "data to write to the source")
	sourcesWriteCmd.MarkFlagRequired("id")
	sourcesWriteCmd.MarkFlagRequired("data")

	sourcesCmd.PersistentFlags().String("host", "localhost", "the host sources server")

	sourcesCmd.AddCommand(sourcesReadCmd)
	sourcesCmd.AddCommand(sourcesWriteCmd)
	rootCmd.AddCommand(sourcesCmd)
}
