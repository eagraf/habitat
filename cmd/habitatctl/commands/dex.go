package commands

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/eagraf/habitat/structs/ctl"
	"github.com/spf13/cobra"
)

var dexCmd = &cobra.Command{
	Use:   "dex",
	Short: "Dex commands",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

var schemaCmd = &cobra.Command{
	Use:   "schema <schema_hash>",
	Short: "Read schema",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			printError(errors.New("Not enough arguments - need schema hash"))
		}

		req := &ctl.SchemaRequest{
			Hash: args[0],
		}
		var res ctl.SchemaResponse
		postRequest(ctl.CommandDexSchemaGet, req, &res)
		// Pretty print schema
		out, err := json.MarshalIndent(&res.Schema, "", "    ")
		if err != nil {
			printError(err)
		}
		fmt.Println(string(out))
	},
}

var schemaWriteCmd = &cobra.Command{
	Use:   "write [-f file_path] | <schema_data>",
	Short: " Write schema",
	Run: func(cmd *cobra.Command, args []string) {
		var data []byte
		fileFlag := cmd.Flags().Lookup("file")
		if fileFlag != nil && fileFlag.Changed {
			buf, err := os.ReadFile(fileFlag.Value.String())
			if err != nil {
				printError(err)
			}
			data = buf
		} else {
			if len(args) < 1 {
				printError(errors.New("not enough arguments - need schema data"))
			}
			data = []byte(args[0])
		}

		// TODO validate JSON schema here
		req := &ctl.SchemaWriteRequest{
			Schema: data,
		}
		var res ctl.SchemaWriteResponse
		postRequest(ctl.CommandDexSchemaWrite, req, &res)
		pretty, err := json.MarshalIndent(res, "", "    ")
		if err != nil {
			printError(err)
		}
		fmt.Println(string(pretty))
	},
}

var interfaceCmd = &cobra.Command{
	Use:   "interface <interface_hash>",
	Short: "Read interface",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			printError(errors.New("not enough arguments - need schema hash"))
		}

		req := &ctl.InterfaceRequest{
			Hash: args[0],
		}
		var res ctl.InterfaceResponse
		postRequest(ctl.CommandDexInterfaceGet, req, &res)
		// Pretty print schema
		out, err := json.MarshalIndent(&res.Interface, "", "    ")
		if err != nil {
			printError(err)
		}
		fmt.Println(string(out))
	},
}

var interfaceWriteCmd = &cobra.Command{
	Use:   "write [-f file_path] | <schema_data>",
	Short: " Write interface",
	Run: func(cmd *cobra.Command, args []string) {
		var data []byte
		fileFlag := cmd.Flags().Lookup("file")
		if fileFlag != nil && fileFlag.Changed {
			buf, err := os.ReadFile(fileFlag.Value.String())
			if err != nil {
				printError(err)
			}
			data = buf
		} else {
			if len(args) < 1 {
				printError(errors.New("not enough arguments - need interface data"))
			}
			data = []byte(args[0])
		}

		var iface ctl.Interface
		err := json.Unmarshal(data, &iface)
		if err != nil {
			printError(err)
		}

		// TODO validate JSON schema here
		req := &ctl.InterfaceWriteRequest{
			Interface: &iface,
		}
		var res ctl.InterfaceWriteResponse
		postRequest(ctl.CommandDexInterfaceWrite, req, &res)
		pretty, err := json.MarshalIndent(res, "", "    ")
		if err != nil {
			printError(err)
		}
		fmt.Println(string(pretty))
	},
}

func init() {
	schemaWriteCmd.Flags().StringP("file", "f", "", "file containing JSON schema")
	interfaceWriteCmd.Flags().StringP("file", "f", "", "file containing interface def")

	schemaCmd.AddCommand(schemaWriteCmd)
	interfaceCmd.AddCommand(interfaceWriteCmd)

	dexCmd.AddCommand(schemaCmd)
	dexCmd.AddCommand(interfaceCmd)

	rootCmd.AddCommand(dexCmd)
}
