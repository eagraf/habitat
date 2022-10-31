package configuration

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/eagraf/habitat/pkg/compass"
	"gopkg.in/yaml.v3"
)

func ReadAppConfig(path string) (*App, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("error opening %s: %s", path, err)
	}
	defer file.Close()

	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}

	var res App
	err = yaml.Unmarshal(bytes, &res)
	if err != nil {
		return nil, err
	}

	return &res, nil
}

func GetAppConfig(appName string) (*App, string, error) {

	path, err := compass.FindAppPath(appName)
	if err != nil {
		return nil, "", err
	}

	configPath := filepath.Join(path, "habitat.yaml")

	config, err := ReadAppConfig(configPath)
	if err != nil {
		return nil, path, err
	}
	return config, path, nil
}
