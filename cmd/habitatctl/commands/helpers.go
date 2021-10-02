package commands

import (
	"fmt"

	"github.com/eagraf/habitat/cmd/habitatctl/client"
	"github.com/eagraf/habitat/structs/ctl"
)

func sendRequest(command string, args []string) {
	client, err := client.NewClient()
	if err != nil {
		fmt.Println("Error: couldn't connect to habitat service")
		return
	}

	client.WriteRequest(&ctl.Request{
		Command: command,
		Args:    args,
	})

	res, err := client.ReadResponse()
	if err != nil {
		fmt.Println("Error: couldn't read response from habitat service")
	}
	fmt.Println(res)
}
