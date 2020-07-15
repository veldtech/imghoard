package config

import (
	"encoding/json"
	"io/ioutil"
)

func LoadFromFile(uri string) (Config, error) {
	fileBuffer, err := ioutil.ReadFile(uri)
	if err != nil {
		return Config{}, err
	}

	config := Config{}
	err = json.Unmarshal(fileBuffer, &config)
	if err != nil {
		return Config{}, err
	}
	return config, nil
}