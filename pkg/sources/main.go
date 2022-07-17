package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/eagraf/habitat/pkg/sources/sources"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var reader sources.Reader
var writer sources.Writer

var rootCmd = &cobra.Command{
	Use:   "sources",
	Short: "Sources allows reading and writing to local data",
	Long:  `Sources allows reading and writinng to local data`,
	Run: func(cmd *cobra.Command, args []string) {
		// Do Stuff Here
		fmt.Println("root cmd")
	},
}

var readCmd = &cobra.Command{
	Use:   "read",
	Short: "read a source file",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			log.Error("Not enough arguments to read command, needs source name")
		}

		req := &sources.ReadRequest{}
		json.Unmarshal([]byte(args[1]), req)
		allowed, err, data := reader.Read(*req)

		fmt.Sprintf("Source Read Request from app %s for: %s\n", req.Requester, req.Source.Name)
		if err != nil {
			log.Error("Error reading source: ", err.Error())
		} else {
			fmt.Sprintf("Allowed: %t, Data: %s", allowed, string(data))
		}
	},
}

var writeCmd = &cobra.Command{
	Use:   "write",
	Short: "write a source file",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			log.Error("Not enough arguments to read command, needs source name")
		}

		req := &sources.WriteRequest{}
		json.Unmarshal([]byte(args[1]), req)
		allowed, err := writer.Write(*req)

		fmt.Sprintf("Source Write Request from app %s for: %s\n", req.Requester, req.Source.Name)
		if err != nil {
			log.Error("Error reading source: ", err.Error())
		} else {
			fmt.Sprintf("Allowed: %t, Data: %s", allowed, string(req.Data))
		}
	},
}

func Execute(sreader sources.Reader, swriter sources.Writer) {
	reader = sreader
	writer = swriter
	rootCmd.AddCommand(readCmd)
	rootCmd.AddCommand(writeCmd)
	rootCmd.Execute()
}

func main() {
	if len(os.Args) < 2 {
		log.Error("Not enough arguments, enter sources directory")
		os.Exit(1)
	}
	basePath := "~/Desktop/sources"
	sreader := sources.NewJSONReader(basePath)
	swriter := sources.NewJSONWriter(basePath)
	pmanager := sources.NewBasicPermissionsManager()
	reader := sources.NewReader(sreader, pmanager)
	writer := sources.NewWriter(swriter, pmanager)

	Execute(*reader, *writer)
}
