package dex

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

// Common code that can be used by all DEX clients written in Go
// dex-{datastore} client schema <hash>
// dex-{datastore} client schema -w <schema_bytes>
// dex-{datastore} client schema -w -f <schema_file>

type ClientCmdFactory struct {
	getWriteClientFunc GetWriteClientFunc
	//	client             *Client
}

func (c *ClientCmdFactory) getClient(cmd *cobra.Command) *Client {
	writeClient, err := c.getWriteClientFunc(cmd)
	if err != nil {
		log.Fatal().Err(err)
	}

	sockPath, err := cmd.Flags().GetString("socket-path")
	if err != nil {
		log.Fatal().Err(err)
	}

	client, err := NewClient(sockPath, writeClient)
	if err != nil {
		log.Fatal().Err(err)
	}
	return client
}

type GetWriteClientFunc func(cmd *cobra.Command) (WriteClient, error)

func NewClientCmdFactory(writeClientFunc GetWriteClientFunc) *ClientCmdFactory {
	return &ClientCmdFactory{
		getWriteClientFunc: writeClientFunc,
	}
}

func errorAndExit(msg string) {
	fmt.Println(msg)
	os.Exit(1)
}

func (c *ClientCmdFactory) SchemaCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "schema <schema_hash>",
		Short: "Read schema",
		Run: func(cmd *cobra.Command, args []string) {
			client := c.getClient(cmd)
			if len(args) < 1 {
				errorAndExit("Not enough arguments - need schema hash")
			}
			res, err := client.Schema(args[0])
			if err != nil {
				errorAndExit(err.Error())
			}
			// Pretty print schema
			out, err := json.MarshalIndent(&res.Schema, "", "    ")
			if err != nil {
				errorAndExit(err.Error())
			}
			fmt.Println(string(out))
		},
	}

	writeCmd := &cobra.Command{
		Use:   "write [-f file_path] | <schema_data>",
		Short: " Write schema",
		Run: func(cmd *cobra.Command, args []string) {
			client := c.getClient(cmd)
			if client.WriteClient == nil {
				fmt.Println("No write client implemented, exiting")
				return
			}
			var data []byte
			fileFlag := cmd.Flags().Lookup("file")
			if fileFlag != nil && fileFlag.Changed {
				buf, err := os.ReadFile(fileFlag.Value.String())
				if err != nil {
					errorAndExit(err.Error())
				}
				data = buf
			} else {
				data = []byte(args[0])
			}

			// TODO validate JSON schema here

			hash, err := Schema(data).Hash()
			if err != nil {
				errorAndExit(err.Error())
			}

			res, err := client.WriteSchema(hash, Schema(data))
			if err != nil {
				errorAndExit(err.Error())
			}
			pretty, err := json.MarshalIndent(res, "", "    ")
			if err != nil {
				errorAndExit(err.Error())
			}
			fmt.Println(string(pretty))
		},
	}

	//cmd.PersistentFlags().StringP("socket-path", "s", "", "UNIX socket path")
	writeCmd.Flags().StringP("file", "f", "", "file containing JSON schema")

	cmd.AddCommand(writeCmd)
	return cmd
}

func (c *ClientCmdFactory) InterfaceCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "interface <interface_hash>>",
		Short: "Read interface",
		Run: func(cmd *cobra.Command, args []string) {
			client := c.getClient(cmd)
			if len(args) < 1 {
				errorAndExit("Not enough arguments - need interface hash")
			}
			res, err := client.Interface(args[0])
			if err != nil {
				errorAndExit(err.Error())
			}
			// Pretty print schema
			out, err := json.MarshalIndent(res.Interface, "", "    ")
			if err != nil {
				errorAndExit(err.Error())
			}
			fmt.Println(string(out))
		},
	}

	writeCmd := &cobra.Command{
		Use:   "write [-f file_path] | <interface_data>",
		Short: " Write interface",
		Run: func(cmd *cobra.Command, args []string) {
			client := c.getClient(cmd)
			if client.WriteClient == nil {
				fmt.Println("No write client implemented, exiting")
				return
			}

			var data []byte
			fileFlag := cmd.Flags().Lookup("file")
			if fileFlag != nil && fileFlag.Changed {
				buf, err := os.ReadFile(fileFlag.Value.String())
				if err != nil {
					errorAndExit(err.Error())
				}
				data = buf
			} else {
				data = []byte(args[0])
			}

			// TODO validate interface here
			// TODO make sure the schema can be found

			var iface Interface
			err := json.Unmarshal(data, &iface)
			if err != nil {
				errorAndExit(err.Error())
			}

			hash, err := iface.Hash()
			if err != nil {
				errorAndExit(err.Error())
			}

			res, err := client.WriteInterface(hash, &iface)
			if err != nil {
				errorAndExit(err.Error())
			}

			pretty, err := json.MarshalIndent(res, "", "    ")
			if err != nil {
				errorAndExit(err.Error())
			}
			fmt.Println(string(pretty))
		},
	}

	//	cmd.PersistentFlags().StringP("socket-path", "s", "", "UNIX socket path")
	writeCmd.Flags().StringP("file", "f", "", "file containing JSON schema")

	cmd.AddCommand(writeCmd)
	return cmd
}

func (c *ClientCmdFactory) ImplementationsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "implementations <interface_hash>",
		Short: "Read implementations",
		Run: func(cmd *cobra.Command, args []string) {
			client := c.getClient(cmd)

			fmt.Println("reading")
			if len(args) < 1 {
				errorAndExit("Not enough arguments - need interface hash")
			}
			res, err := client.Implementations(args[0])
			if err != nil {
				errorAndExit(err.Error())
			}

			out, err := json.MarshalIndent(res.Implementations, "", "    ")
			if err != nil {
				errorAndExit(err.Error())
			}
			fmt.Println(string(out))
		},
	}

	addCmd := &cobra.Command{
		Use:   "add <interface_hash> <datastore_id> <query_string>",
		Short: "Add an implementation",
		Run: func(cmd *cobra.Command, args []string) {
			client := c.getClient(cmd)
			if client.WriteClient == nil {
				fmt.Println("No write client implemented, exiting")
				return
			}
			fmt.Println("adding")

			if len(args) < 3 {
				errorAndExit("Not enough arguments - need interface hash")
			}

			err := client.AddImplementation(args[0], args[1], args[2])
			if err != nil {
				errorAndExit(err.Error())
			}
		},
	}

	removeCmd := &cobra.Command{
		Use:   "remove <interface_hash> <datastore_id> <query_string>",
		Short: "remove an implementation",
		Run: func(cmd *cobra.Command, args []string) {
			client := c.getClient(cmd)
			if client.WriteClient == nil {
				fmt.Println("No write client implemented, exiting")
				return
			}

			if len(args) < 2 {
				errorAndExit("Not enough arguments - need interface hash")
			}

			err := client.RemoveImplementation(args[0], args[1])
			if err != nil {
				errorAndExit(err.Error())
			}
		},
	}

	//	cmd.PersistentFlags().StringP("socket-path", "s", "", "UNIX socket path")

	cmd.AddCommand(addCmd)
	cmd.AddCommand(removeCmd)

	return cmd
}

/*func (c *ClientCmd) WriteSchemaCmd() *cobra.Command {
}

func (c *ClientCmd) WriteInterfaceCmd() *cobra.Command {
}

func (c *ClientCmd) WriteImplementationCmd() *cobra.Command {
}*/
