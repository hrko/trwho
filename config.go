package main

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"

	"github.com/adrg/xdg"
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

func SearchConfigFile() (string, error) {
	configDirs := make([]string, 0)
	configDirs = append(configDirs, xdg.ConfigHome)
	configDirs = append(configDirs, "/etc")
	for _, dir := range configDirs {
		configPath := filepath.Join(dir, appName, "config.json")
		if _, err := os.Stat(configPath); err == nil {
			return configPath, nil
		}
	}
	return "", errors.New("Config file not found")
}
