package commands

import (
	"encoding/json"
	"fmt"

	"github.com/eagraf/habitat/cmd/sources"
	"github.com/eagraf/habitat/pkg/compass"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

// communityCmd represents the community command
var sourcesCmd = &cobra.Command{
	Use:   "sources",
	Short: "Habitat sources is the data access layer on top of JSON, relational, and key-value data.",
	Long: `Subcommands:
	read <source_file>
	join <source_file>
`,
	Run: func(cmd *cobra.Command, args []string) {
	},
}

var sourcesReadCmd = &cobra.Command{
	Use:   "read",
	Short: "read a source file",
	Run: func(cmd *cobra.Command, args []string) {

		id := cmd.Flags().Lookup("id").Value.String()
		path := cmd.Flags().Lookup("path").Value.String()

		fmt.Printf("Source Read Request for %s\n", id)

		req := &sources.ReadRequest{
			SourceName: sources.SourceName(id),
		}

		reader := sources.NewReader(sources.NewJSONReader(cmd.Context(), path), sources.NewBasicPermissionsManager())

		allowed, err, data := reader.Read(req)

		if err != nil {
			log.Error().Msgf("Error reading source: %s", err.Error())
		} else {
			fmt.Printf("Allowed: %t, Data: %s\n", allowed, data)
		}
	},
}

var sourcesWriteCmd = &cobra.Command{
	Use:   "write",
	Short: "write a source file",
	Run: func(cmd *cobra.Command, args []string) {

		id := cmd.Flags().Lookup("id").Value.String()
		path := cmd.Flags().Lookup("path").Value.String()
		data := cmd.Flags().Lookup("data").Value.String()

		fmt.Printf("Source Write Request for %s, %s\n", id, data)

		req := &sources.WriteRequest{
			SourceName: sources.SourceName(id),
			Data:       json.RawMessage(data),
		}

		writer := sources.NewWriter(sources.NewJSONWriter(cmd.Context(), path), sources.NewBasicPermissionsManager())

		allowed, err := writer.Write(req)

		if err != nil {
			log.Error().Msgf("Error reading source: %s", err.Error())
		} else {
			fmt.Printf("Allowed: %t, Data: %s\n", allowed, data)
		}
	},
}

/*
var sourcesAddCmd = &cobra.Command{
	Use:   "add",
	Short: "add a source to local registry",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			log.Error().Msg("Not enough arguments to read command, needs source name")
			return
		}

		input, err := os.ReadFile(args[0])
		if err != nil {
			log.Error().Msgf("Error reading input: %s", err.Error())
			return
		}
		req := &sources.WriteRequest{}
		err = json.Unmarshal(input, req)
		if err != nil {
			log.Error().Msgf("Erroring unmarshaling json: %s", err.Error())
			return
		}
		allowed, err := writer.Write(req)

		fmt.Printf("Source Write Request from app %s for: %s\n", req.Requester, req.SourceName)
		if err != nil {
			log.Error().Msgf("Error reading source: %s", err.Error())
		} else {
			fmt.Printf("Allowed: %t, Data: %s\n", allowed, req.Data)
		}
	},
}

var sourcesDeleteCmd = &cobra.Command{
	Use:   "remove",
	Short: "remove a source from the local registry",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			log.Error().Msg("Not enough arguments to read command, needs source name")
			return
		}

		input, err := os.ReadFile(args[0])
		if err != nil {
			log.Error().Msgf("Error reading input: %s", err.Error())
			return
		}
		req := &sources.WriteRequest{}
		err = json.Unmarshal(input, req)
		if err != nil {
			log.Error().Msgf("Erroring unmarshaling json: %s", err.Error())
			return
		}
		allowed, err := writer.Write(req)

		fmt.Printf("Source Write Request from app %s for: %s\n", req.Requester, req.SourceName)
		if err != nil {
			log.Error().Msgf("Error reading source: %s", err.Error())
		} else {
			fmt.Printf("Allowed: %t, Data: %s\n", allowed, req.Data)
		}
	},
}
*/

func init() {

	sourcesReadCmd.Flags().String("id", "", "id (name) of the source being read")
	sourcesReadCmd.MarkFlagRequired("id")
	sourcesWriteCmd.Flags().String("id", "", "id (name) of the source being read")
	sourcesWriteCmd.Flags().StringP("data", "d", "", "data to write to the source")
	sourcesWriteCmd.MarkFlagRequired("id")
	sourcesWriteCmd.MarkFlagRequired("data")

	sourcesCmd.PersistentFlags().String("path", compass.LocalSourcesPath(), "The path where sources data is located")
	sourcesCmd.MarkFlagRequired("path")

	sourcesCmd.AddCommand(sourcesReadCmd)
	sourcesCmd.AddCommand(sourcesWriteCmd)
	rootCmd.AddCommand(sourcesCmd)
}
