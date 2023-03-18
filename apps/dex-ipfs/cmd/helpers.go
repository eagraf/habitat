package cmd

import "fmt"

func schemaPath(hash string) string {
	return fmt.Sprintf("/schema/%s", hash)
}

func interfacePath(hash string) string {
	return fmt.Sprintf("/interface/%s", hash)
}

func implementationsPath(ifaceHash string) string {
	return fmt.Sprintf("/implementations/%s", ifaceHash)
}
