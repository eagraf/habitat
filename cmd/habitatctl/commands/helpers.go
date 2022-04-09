package commands

import (
	"fmt"
	"os"

	client "github.com/eagraf/habitat/pkg/habitat_client"
	"github.com/eagraf/habitat/structs/ctl"
	"github.com/spf13/viper"
)

func habitatServiceAddr() string {
	return fmt.Sprintf("localhost:%s", viper.GetString("port"))
}

func sendRequest(req interface{}) *ctl.ResponseWrapper {
	res, err := client.SendRequestToAddress(habitatServiceAddr(), req)
	if err != nil {
		printError(err)
	}
	return res
}

func printError(err error) {
	fmt.Printf("failed to make request: %s\n", err)
	os.Exit(1)
}
