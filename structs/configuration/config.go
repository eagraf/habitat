package configuration

import (
	"io/ioutil"
	"os"

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
