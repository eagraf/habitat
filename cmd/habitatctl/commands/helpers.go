package commands

import (
	"fmt"
	"os"

	"github.com/eagraf/habitat/pkg/compass"
	client "github.com/eagraf/habitat/pkg/habitat_client"
	"github.com/eagraf/habitat/structs/ctl"
	"github.com/spf13/viper"
)

func printError(err error) {
	fmt.Printf("failed to make request: %s\n", err)
	os.Exit(1)
}

func postRequest(reqType string, req, res interface{}) {
	if viper.IsSet("libp2p-proxy") {
		proxyAddr := viper.GetString("libp2p-proxy")

		err, apiErr := client.PostLibP2PRequestToAddress(proxyAddr, ctl.GetRoute(reqType), req, res)
		if err != nil {
			printError(fmt.Errorf("error submitting request: %s", err))
		} else if apiErr != nil {
			printError(apiErr)
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
