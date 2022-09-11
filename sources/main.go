package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path"

	"github.com/eagraf/habitat/pkg/compass"
	"github.com/eagraf/habitat/sources/sources"
	"github.com/eagraf/habitat/utils"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var reader *sources.Reader
var writer *sources.Writer

var rootCmd = &cobra.Command{
	Use:   "sources",
	Short: "Sources allows reading and writing to local data",
	Long:  `Sources allows reading and writinng to local data`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Run at beginning
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
	},
}

var readCmd = &cobra.Command{
	Use:   "read",
	Short: "read a source file",
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
		req := &sources.ReadRequest{}
		err = json.Unmarshal(input, req)
		if err != nil {
			log.Error().Msgf("Erroring unmarshaling json: %s", err.Error())
			return
		}
		allowed, err, data := reader.Read(req)

		fmt.Printf("Source Read Request from app %s for: %s\n", req.Requester, req.Source.Name)
		if err != nil {
			log.Error().Msgf("Error reading source: %s", err.Error())
		} else {
			fmt.Printf("Allowed: %t, Data: %s\n", allowed, data)
		}
	},
}

var writeCmd = &cobra.Command{
	Use:   "write",
	Short: "write a source file",
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

		fmt.Printf("Source Write Request from app %s for: %s\n", req.Requester, req.Source.Name)
		if err != nil {
			log.Error().Msgf("Error reading source: %s", err.Error())
		} else {
			fmt.Printf("Allowed: %t, Data: %s\n", allowed, req.Data)
		}
	},
}

func Execute() {
	rootCmd.AddCommand(readCmd)
	rootCmd.AddCommand(writeCmd)
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func main() {
	sourcesPath := utils.GetEnvDefault("SOURCES_PATH", "~/habitat/data/sources")
	sreader := sources.NewJSONReader(sourcesPath)
	swriter := sources.NewJSONWriter(sourcesPath)
	pmanager := sources.NewBasicPermissionsManager()
	reader = sources.NewReader(sreader, pmanager)
	writer = sources.NewWriter(swriter, pmanager)

	rootCmd.PersistentFlags().String("path", path.Join(compass.HabitatPath(), "sources"), "The path where sources data is located")
	rootCmd.MarkFlagRequired("path")
	Execute()

	// TODO: how do we get data server nodes?
	sourcesServer := sources.NewSourcesServer(reader, writer, map[string]sources.DataServerNode{})
	sourcesServer.Start()
}
