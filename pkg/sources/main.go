package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/eagraf/habitat/pkg/sources/sources"
	log "github.com/sirupsen/logrus"
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
		basePath, _ := cmd.Flags().GetString("path")
		sreader := sources.NewJSONReader(basePath)
		swriter := sources.NewJSONWriter(basePath)
		pmanager := sources.NewBasicPermissionsManager()
		reader = sources.NewReader(sreader, pmanager)
		writer = sources.NewWriter(swriter, pmanager)
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
			log.Error("Not enough arguments to read command, needs source name")
			return
		}

		input, err := os.ReadFile(args[0])
		if err != nil {
			log.Error("Error reading input: ", err.Error())
			return
		}
		req := &sources.ReadRequest{}
		err = json.Unmarshal(input, req)
		if err != nil {
			log.Error("Erroring unmarshaling json: ", err.Error())
			return
		}
		allowed, err, data := reader.Read(req)

		fmt.Printf("Source Read Request from app %s for: %s\n", req.Requester, req.Source.Name)
		if err != nil {
			log.Error("Error reading source: ", err.Error())
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
			log.Error("Not enough arguments to read command, needs source name")
			return
		}

		input, err := os.ReadFile(args[0])
		if err != nil {
			log.Error("Error reading input: ", err.Error())
			return
		}
		req := &sources.WriteRequest{}
		err = json.Unmarshal(input, req)
		if err != nil {
			log.Error("Erroring unmarshaling json: ", err.Error())
			return
		}
		allowed, err := writer.Write(req)

		fmt.Printf("Source Write Request from app %s for: %s\n", req.Requester, req.Source.Name)
		if err != nil {
			log.Error("Error reading source: ", err.Error())
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
	rootCmd.PersistentFlags().String("path", "~/Desktop/sources", "The path where sources data is located")
	Execute()
}
