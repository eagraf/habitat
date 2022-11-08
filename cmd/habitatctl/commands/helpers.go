package commands

import (
	"fmt"
	"os"

	client "github.com/eagraf/habitat/pkg/habitat_client"
	"github.com/eagraf/habitat/structs/ctl"
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
	err, apiErr := client.PostRequestToAddress(habitatServiceAddr()+ctl.GetRoute(reqType), req, res)
	if err != nil {
		printError(fmt.Errorf("error submitting request: %s", err))
	} else if apiErr != nil {
		printError(apiErr)
	}
}
