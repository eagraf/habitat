package commands

import (
	"fmt"

	client "github.com/eagraf/habitat/pkg/habitat_client"
	"github.com/spf13/viper"
)

func habitatServiceAddr() string {
	return fmt.Sprintf("localhost:%s", viper.GetString("port"))
}

func SendRequestAndPrint(command string, args []string) {
	res, err := client.SendRequestToAddress(habitatServiceAddr(), command, args)
	if err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Println(res)
	}
}
