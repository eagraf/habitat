package commands

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/eagraf/habitat/pkg/compass"
	"github.com/eagraf/habitat/pkg/identity"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

var idCmd = &cobra.Command{
	Use:   "id",
	Short: "Work with Habitat IDs",
	Long: `Subcommands:
	create
	`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(cmd.Usage())
	},
}

var idInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize habitatctl identity management",
	Long:  "habitatctl checks the value of the HABITATCTL_IDENTITY_PATH env variable. If no value is found, the default for your operating system is used.",
	Run: func(cmd *cobra.Command, args []string) {

		fmt.Println("Initializing Habitat identity management")

		identityPath := compass.HabitatIdentityPath()
		err := os.MkdirAll(identityPath, 0700)
		if err != nil {
			fmt.Printf("error creating identity directory %s: %s\n", identityPath, err)
			os.Exit(1)
		}

		fmt.Println("Success!")
	},
}

var idListCmd = &cobra.Command{
	Use:   "ls",
	Short: "List all identities managed by Habitat",
	Run: func(cmd *cobra.Command, args []string) {
		identityPath := checkIdentityPath()
		files, err := ioutil.ReadDir(identityPath)
		if err != nil {
			fmt.Printf("error reading identity files: %s", err)
			os.Exit(1)
		}
		for _, f := range files {
			if strings.HasSuffix(f.Name(), ".cert") {
				fmt.Println(strings.TrimSuffix(f.Name(), ".cert"))
			}
		}
	},
}

var idCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Generate a new Habitat ID",
	Run: func(cmd *cobra.Command, args []string) {

		identityPath := checkIdentityPath()

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

		uuid := uuid.New().String()
		fmt.Printf("generating keypair for username %s with UUID %s\n", username, uuid)

		user, err := identity.GenerateNewUserCert(username, uuid)
		if err != nil {
			fmt.Printf("error generating new user certifiate: %s\n", err)
			os.Exit(1)
		}

		err = identity.StoreUserIdentity(identityPath, user, []byte(password))
		if err != nil {
			fmt.Printf("error storing new user identity: %s\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	idCreateCmd.Flags().String("password", "", "password used to encrypt the private key file")
	idCreateCmd.Flags().StringP("username", "u", "", "username for new identity")

	idCmd.AddCommand(idInitCmd)
	idCmd.AddCommand(idListCmd)
	idCmd.AddCommand(idCreateCmd)
	rootCmd.AddCommand(idCmd)
}
