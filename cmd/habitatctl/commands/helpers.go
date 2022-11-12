package commands

import (
	"errors"
	"fmt"
	"os"

	"github.com/eagraf/habitat/pkg/compass"
	client "github.com/eagraf/habitat/pkg/habitat_client"
	"github.com/eagraf/habitat/pkg/identity"
	"github.com/eagraf/habitat/structs/ctl"
	"github.com/gorilla/websocket"
	"github.com/libp2p/go-libp2p/core/peer"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func habitatServiceAddr() string {
	return fmt.Sprintf("http://localhost:%s", viper.GetString("port"))
}

func printError(err error) {
	fmt.Printf("failed to make request: %s\n", err)
	os.Exit(1)
}

func postRequest(reqType string, req, res interface{}) {
	if viper.IsSet("libp2p-proxy") {
		proxyAddr := viper.GetString("libp2p-proxy")

		remoteMA, err := ma.NewMultiaddr(proxyAddr)
		if err != nil {
			printError(err)
		}

		b58PeerID, err := remoteMA.ValueForProtocol(ma.P_P2P)
		if err != nil {
			printError(err)
		}

		addrStr := ""
		ip, err := remoteMA.ValueForProtocol(ma.P_IP4)
		if err != nil {
			ip, err = remoteMA.ValueForProtocol(ma.P_IP6)
			if err != nil {
				printError(errors.New("supplied libp2p multiaddr does not contain an IP address"))
			} else {
				addrStr += "/ip6/" + ip
			}
		} else {
			addrStr += "/ip4/" + ip
		}

		port, err := remoteMA.ValueForProtocol(ma.P_TCP)
		if err != nil {
			printError(err)
		}
		addrStr += "/tcp/" + port

		addr, err := ma.NewMultiaddr(addrStr)
		if err != nil {
			printError(err)
		}

		// decode base58 encoded peer id for setting addresses
		peerID, err := peer.Decode(b58PeerID)
		if err != nil {
			printError(err)
		}

		err, apiErr := client.PostLibP2PRequestToAddress(addr, ctl.GetRoute(reqType), peerID, req, res)
		if err != nil {
			printError(fmt.Errorf("error submitting request: %s", err))
		} else if apiErr != nil {
			printError(apiErr)
		}
	} else {
		err, apiErr := client.PostRequestToAddress(habitatServiceAddr()+ctl.GetRoute(reqType), req, res)
		if err != nil {
			printError(fmt.Errorf("error submitting request: %s", err))
		} else if apiErr != nil {
			printError(apiErr)
		}
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
		fmt.Printf("error loading user identity for %s\n", username)
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
	return client.GetWebsocketConn(fmt.Sprintf("ws://localhost:%s%s", viper.GetString("port"), ctl.GetRoute(reqType)))
}
