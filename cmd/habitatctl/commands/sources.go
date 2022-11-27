package commands

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/eagraf/habitat/cmd/sources"
	"github.com/eagraf/habitat/pkg/compass"
	"github.com/qri-io/jsonschema"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var rw *sources.JSONReaderWriter
var sr *sources.SchemaRegistry

// communityCmd represents the community command
var sourcesCmd = &cobra.Command{
	Use:   "sources",
	Short: "Habitat sources is the data access layer on top of JSON, relational, and key-value data.",
	Long: `Subcommands:
	read <source_file>
	join <source_file>
`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		rw = sources.NewJSONReaderWriter(cmd.Context(), cmd.Flags().Lookup("sources_path").Value.String())
		sr = sources.NewSchemaRegistry(cmd.Flags().Lookup("schema_path").Value.String())
	},
}

var sourcesReadCmd = &cobra.Command{
	Use:   "read",
	Short: "read a source file",
	Run: func(cmd *cobra.Command, args []string) {

		id := cmd.Flags().Lookup("id").Value.String()

		fmt.Printf("Source Read Request for %s\n", id)

		data, err := rw.Read(sources.SourceID(id))

		if err != nil {
			log.Error().Msgf("Error reading source: %s", err.Error())
		} else {
			fmt.Printf("Data: %s\n", data)
		}
	},
}

var sourcesWriteCmd = &cobra.Command{
	Use:   "write",
	Short: "write a source file",
	Run: func(cmd *cobra.Command, args []string) {

		id := cmd.Flags().Lookup("id").Value.String()
		data := cmd.Flags().Lookup("data").Value.String()

		fmt.Printf("Source Write Request for %s, %s\n", id, data)

		sch, err := sr.Lookup(id)
		if err != nil || sch == nil {
			fmt.Printf("Did not find schema in registry: %s", err)
			return
		}

		err = rw.Write(sources.SourceID(id), sch, []byte(data))
		if err != nil {
			log.Error().Msgf("Error writing source: %s", err.Error())
		} else {
			fmt.Printf("Wrote Data: %s\n", data)
		}
	},
}

var sourcesAddCmd = &cobra.Command{
	Use:   "add",
	Short: "add a source schema to local registry",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			log.Error().Msg("not enough arguments to sources add, needs file with schema")
			return
		}

		input, err := os.ReadFile(args[0])
		if err != nil {
			log.Error().Msgf("Error reading input: %s", err.Error())
			return
		}

		sch := jsonschema.Must(string(input))
		id := sources.GetSchemaId(sch)
		fmt.Println(sr.Add(id, sch))
	},
}

var sourcesShowCmd = &cobra.Command{
	Use:   "show",
	Short: "show everything in local registry",
	Run: func(cmd *cobra.Command, args []string) {

		path := sr.Path
		files, err := ioutil.ReadDir(path)
		if err != nil {
			fmt.Printf("error reading path %s: %s", path, err.Error())
			return
		}
		for _, file := range files {
			fmt.Println(file.Name(), file.IsDir())
		}
	},
}

/*
var sourcesDeleteCmd = &cobra.Command{
	Use:   "remove",
	Short: "remove a source from the local registry",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			log.Error().Msg("not enough arguments to sources add, needs file with schema")
			return
		}

		input, err := os.ReadFile(args[0])
		if err != nil {
			log.Error().Msgf("Error reading input: %s", err.Error())
			return
		}

		sch := jsonschema.Must(string(input))
		sr.Delete(sources.GetSchemaId(sch))
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

	sourcesCmd.PersistentFlags().String("sources_path", compass.LocalSourcesPath(), "The path where sources data is located")
	sourcesCmd.MarkFlagRequired("sources_path")

	sourcesCmd.PersistentFlags().String("schema_path", compass.LocalSchemaPath(), "The path where sources data is located")
	sourcesCmd.MarkFlagRequired("schema_path")

	sourcesCmd.AddCommand(sourcesReadCmd)
	sourcesCmd.AddCommand(sourcesWriteCmd)
	sourcesCmd.AddCommand(sourcesAddCmd)
	sourcesCmd.AddCommand(sourcesShowCmd)
	rootCmd.AddCommand(sourcesCmd)
}
