package main

import (
	"encoding/json"
	"os"
)

type Config struct {
	Hosts []*ConfigHostEntry `json:"hosts"`
}

type ConfigHostEntry struct {
	Hostname string `json:"hostname"`
	Note     string `json:"note"`
}

func ReadConfig(path string) (*Config, error) {
	c := new(Config)
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	d := json.NewDecoder(f)
	if err := d.Decode(c); err != nil {
		return nil, err
	}
	return c, nil
}

// func SearchConfigFile() (string, error) {

// }
