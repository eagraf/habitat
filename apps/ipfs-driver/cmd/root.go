/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/eagraf/habitat/pkg/ipfs"
	"github.com/eagraf/habitat/structs/community"
	"github.com/spf13/cobra"
)

func errorAndExit(err error) {
	fmt.Printf("%s\n", err)
	os.Exit(1)
}

// TODO this should use the go-multiaddr library
func changeMultiaddrPort(addr string, port string) string {
	parts := strings.Split(addr, "/")
	parts[len(parts)-1] = port
	return strings.Join(parts, "/")
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "ipfs-driver",
	Short: "A brief description of your application",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			errorAndExit(errors.New("no IPFS path specified"))
		}

		ipfsPath := args[0]

		ipfsInstance := &ipfs.IPFSInstance{
			IPFSPath: ipfsPath,
		}
		// check if IPFS_PATH exists
		_, err := os.Stat(ipfsPath)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				// 	if not initialize
				err := ipfsInstance.Init()
				if err != nil {
					errorAndExit(err)
				}
			} else {
				errorAndExit(fmt.Errorf("error running stat on %s: %s\n", ipfsPath, err))
			}
		} else {
			// check if version file exists
			versionPath := filepath.Join(ipfsPath, "version")
			_, err := os.Stat(versionPath)
			if err != nil {
				if errors.Is(err, os.ErrNotExist) {
					// 	if not initialize
					err := ipfsInstance.Init()
					if err != nil {
						errorAndExit(err)
					}
				} else {
					errorAndExit(fmt.Errorf("error running stat on %s: %s\n", versionPath, err))
				}
			}
		}

		configChanged := false
		config, err := ipfsInstance.Config()
		if err != nil {
			errorAndExit(err)
		}

		flagSet := cmd.Flags()
		// reconfigure
		communityConfigurationB64, err := flagSet.GetString("community-configuration-b64")
		if err != nil {
			errorAndExit(err)
		}
		if communityConfigurationB64 != "" {
			configChanged = true

			communityConfigurationMarshaled, err := base64.StdEncoding.DecodeString(communityConfigurationB64)
			if err != nil {
				errorAndExit(err)
			}

			var communityConfiguration community.IPFSConfig
			err = json.Unmarshal(communityConfigurationMarshaled, &communityConfiguration)
			if err != nil {
				errorAndExit(err)
			}

			config.Bootstrap = communityConfiguration.BootstrapAddresses
			err = ipfsInstance.WriteSwarmKey([]byte(communityConfiguration.SwarmKey))
			if err != nil {
				errorAndExit(err)
			}
		}

		// TODO add some validation for these ports
		swarmPort, err := flagSet.GetString("swarm-port")
		if err != nil {
			errorAndExit(err)
		}
		if swarmPort != "" {
			configChanged = true
			config.Addresses.Swarm = []string{
				changeMultiaddrPort(config.Addresses.Swarm[0], swarmPort),
			}
		}

		apiPort, err := flagSet.GetString("api-port")
		if err != nil {
			errorAndExit(err)
		}
		if apiPort != "" {
			configChanged = true
			config.Addresses.API = []string{
				changeMultiaddrPort(config.Addresses.API[0], apiPort),
			}
		}

		gatewayPort, err := flagSet.GetString("gateway-port")
		if err != nil {
			errorAndExit(err)
		}
		if gatewayPort != "" {
			configChanged = true
			config.Addresses.Gateway = []string{
				changeMultiaddrPort(config.Addresses.Gateway[0], gatewayPort),
			}
		}

		if configChanged {
			ipfsInstance.Configure(config)
		}

		err = ipfsInstance.Daemon()
		if err != nil {
			errorAndExit(err)
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.ipfs-driver.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	rootCmd.PersistentFlags().StringP("community-configuration-b64", "c", "", "base64 encoded community IPFS configuration")
	rootCmd.PersistentFlags().StringP("swarm-port", "s", "", "IPFS swarm port number")
	rootCmd.PersistentFlags().StringP("api-port", "a", "", "IPFS api port number")
	rootCmd.PersistentFlags().StringP("gateway-port", "g", "", "IPFS gateway port number")
}
