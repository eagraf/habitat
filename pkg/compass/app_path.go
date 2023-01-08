package compass

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// The HABITAT_APP_PATH env variable works similar to the PATH env variable, it gives a
// colon separated list of directories for Habitat to search for apps. The first matching
// path with the corresponding app is returned.

// Habitat apps should be structured like so:
// app-dir
//  |-- my-app-name
//  |     |-- bin (dir)
//  |     |-- web (optional dir)
//  |     |-- config.yml
//  |
//  |-- my-second-app
// ......etc.......

func FindAppPath(appName string) (string, error) {
	habitatAppPath := os.Getenv("HABITAT_APP_PATH")
	paths := strings.Split(habitatAppPath, ":")

	for _, p := range paths {
		potentialAppPath := filepath.Join(p, appName)
		file, err := os.Stat(potentialAppPath)
		if errors.Is(err, os.ErrNotExist) {
			continue
		} else if err != nil {
			return "", err
		}

		if !file.IsDir() {
			return "", fmt.Errorf("found %s in app dir %s, but %s is not a directory", appName, p, appName)
		}

		return potentialAppPath, nil
	}

	return "", fmt.Errorf("%s not found in any HABITAT_APP_PATH directories at %s", appName, habitatAppPath)
}
