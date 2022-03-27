package commands

import (
	"fmt"

	client "github.com/eagraf/habitat/pkg/habitat_client"
)

func SendRequestAndPrint(command string, args []string) {
	res, err := client.SendRequest(command, args)
	if err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Println(res)
	}
}
