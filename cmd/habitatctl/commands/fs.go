package commands

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/eagraf/habitat/pkg/compass"
	client "github.com/eagraf/habitat/pkg/habitat_client"
	"github.com/eagraf/habitat/structs/ctl"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var fsCmd = &cobra.Command{
	Use:   "fs",
	Short: "Access the Habitat file system",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(cmd.Usage())
	},
}

var addFileCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a file, and get back its hash",
	Long:  `TODO create long description`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			printError(fmt.Errorf("no file specified"))
		}

		path, err := filepath.Abs(args[0])
		if err != nil {
			printError(err)
		}

		file, err := os.Open(path)
		if err != nil {
			printError(err)
		}
		defer file.Close()

		var res ctl.AddFileResponse
		postFileRequest(ctl.CommandAddFile, nil, &res, file)

		pretty, err := json.MarshalIndent(res, "", "    ")
		if err != nil {
			printError(err)
		}
		fmt.Println(string(pretty))
	},
}

var getFileCmd = &cobra.Command{
	Use:   "get",
	Short: "Get a file by its hash",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			printError(fmt.Errorf("no hash specified"))
		}

		hash := args[0]

		req := &ctl.GetFileRequest{
			ContentID: hash,
		}

		address := compass.CustomHabitatAPIAddr("localhost", viper.GetString("port")) + ctl.GetRoute(ctl.CommandGetFile)
		r, err, apiErr := client.PostRetrieveFileFromAddress(address, req)
		if err != nil {
			printError(fmt.Errorf("error submitting request: %s", err))
		} else if apiErr != nil {
			printError(apiErr)
		}
		// TODO buffered reading for larger files
		buf, err := io.ReadAll(r)
		if err != nil {
			printError(err)
		}
		fmt.Println(string(buf))
	},
}

func init() {
	fsCmd.AddCommand(addFileCmd)
	fsCmd.AddCommand(getFileCmd)
	rootCmd.AddCommand(fsCmd)
}
