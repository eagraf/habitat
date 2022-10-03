package configuration

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/eagraf/habitat/pkg/compass"
	"gopkg.in/yaml.v3"
)

func ReadAppConfigs(path string) (*AppConfiguration, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}

	var res AppConfiguration
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

	file, err := os.Open(configPath)
	if err != nil {
		return nil, configPath, err
	}
	defer file.Close()

	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, configPath, err
	}

	var res App
	err = yaml.Unmarshal(bytes, &res)
	if err != nil {
		return nil, configPath, err
	}

	return &res, path, nil
}
