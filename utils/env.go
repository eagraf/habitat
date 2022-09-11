package utils

import "os"

func GetEnvDefault(varName string, defaultVal string) string {
	val := os.Getenv(varName)
	if val == "" {
		return defaultVal
	}
	return val
}
