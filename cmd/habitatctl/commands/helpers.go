package commands

import (
	"fmt"

	"github.com/eagraf/habitat/cmd/habitatctl/client"
	"github.com/eagraf/habitat/structs/ctl"
	"github.com/spf13/viper"
)

func sendRequest(command string, args []string) {
	client, err := client.NewClient(viper.GetString("port"))
	if err != nil {
		fmt.Println("Error: couldn't connect to habitat service")
		return
	}

	err = client.WriteRequest(&ctl.Request{
		Command: command,
		Args:    args,
	})
	if err != nil {
		fmt.Printf("Error creating request to habitat service: %s", err)
	}

	res, err := client.ReadResponse()
	if err != nil {
		fmt.Printf("Error: couldn't read response from habitat service: %s\n", err)
	}
	fmt.Println(res)
}
