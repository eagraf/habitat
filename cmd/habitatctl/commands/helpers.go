package commands

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"

	"github.com/eagraf/habitat/pkg/compass"
	client "github.com/eagraf/habitat/pkg/habitat_client"
	"github.com/eagraf/habitat/pkg/identity"
	"github.com/eagraf/habitat/pkg/p2p"
	"github.com/eagraf/habitat/structs/ctl"
	"github.com/gorilla/websocket"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func printError(err error) {
	fmt.Printf("failed to make request: %s\n", err)
	os.Exit(1)
}

func postRequest(reqType string, req, res interface{}) {
	if viper.IsSet("libp2p-proxy") {
		proxyAddr := viper.GetString("libp2p-proxy")

		reqBody, err := json.Marshal(req)
		if err != nil {
			printError(fmt.Errorf("error marshaling POST request body: %s", err))
		}

		p2pReq, err := http.NewRequest("POST", "", bytes.NewReader(reqBody))
		if err != nil {
			printError(fmt.Errorf("error constructing HTTP request: %s", err))
		}

		// TODO: change this path if used outside of community or /habitat commands
		bytes, err := p2p.PostLibP2PRequestToAddress(nil, proxyAddr, "/habitat"+ctl.GetRoute(reqType), p2pReq)
		if err != nil {
			printError(fmt.Errorf("error submitting request: %s", err))
		} else if err := json.Unmarshal(bytes, res); err != nil {
			printError(fmt.Errorf("error unmarshalling response: %s", err.Error()))
		}
	} else {
		err, apiErr := client.PostRequestToAddress(compass.CustomHabitatAPIAddr("localhost", viper.GetString("port"))+ctl.GetRoute(reqType), req, res)
		if err != nil {
			printError(fmt.Errorf("error submitting request: %s", err))
		} else if apiErr != nil {
			printError(apiErr)
		}
	}
}

func postFileRequest(reqType string, req, res interface{}, file *os.File) {
	address := compass.CustomHabitatAPIAddr("localhost", viper.GetString("port")) + ctl.GetRoute(reqType)
	err, apiErr := client.PostFileToAddress(address, &http.Client{}, file, res)
	if err != nil {
		printError(fmt.Errorf("error submitting request: %s", err))
	} else if apiErr != nil {
		printError(apiErr)
	}
}

func loadUserIdentity(cmd *cobra.Command) (*identity.UserIdentity, error) {

	if !cmd.Flags().Lookup("username").Changed {
		fmt.Println("bubble tea")
		os.Exit(1)
	}

	username, err := cmd.Flags().GetString("username")
	if err != nil {
		fmt.Println("error reading username flag")
		os.Exit(1)
	}

	password, err := cmd.Flags().GetString("password")
	if err != nil {
		fmt.Println("error reading password flag")
		os.Exit(1)
	}

	idPath := checkIdentityPath()

	userIdentity, err := identity.LoadUserIdentity(idPath, username, []byte(password))
	if err != nil {
		fmt.Printf("error loading user identity for %s: %s\n", username, err.Error())
		os.Exit(1)
	}

	return userIdentity, nil
}

func checkIdentityPath() string {
	idPath := compass.HabitatIdentityPath()
	_, err := os.Stat(idPath)
	if errors.Is(err, os.ErrNotExist) {
		fmt.Println("Habitat identity management not initialized, run `habitatctl id init` to initialize.")
		os.Exit(1)
	}
	return idPath
}

// If only one of these flags is present when command is run, there will be an error. If neither are present,
// the identity will be determined interactively.
func addUserFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("username", "u", "", "username for the identity you wish to run this command with")
	cmd.Flags().String("password", "", "password for the identity you wish to run this command with")
	cmd.MarkFlagsRequiredTogether("username", "password")
}

func getWebsocketConn(reqType string) (*websocket.Conn, error) {
	return client.GetWebsocketConn(compass.CustomHabitatAPIAddrWebsocket("localhost", viper.GetString("port")), ctl.GetRoute(reqType))
}
