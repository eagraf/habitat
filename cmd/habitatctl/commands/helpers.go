package commands

import (
	"errors"
	"fmt"
	"os"

	client "github.com/eagraf/habitat/pkg/habitat_client"
	"github.com/eagraf/habitat/structs/ctl"
	"github.com/libp2p/go-libp2p/core/peer"
	ma "github.com/multiformats/go-multiaddr"
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
