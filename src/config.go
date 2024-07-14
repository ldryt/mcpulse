package main

import (
	"encoding/base64"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	ListenAddress string `yaml:"listen-address"`

	Version struct {
		Name     string `yaml:"name"`
		Protocol int    `yaml:"protocol"`
	} `yaml:"version"`

	Motds struct {
		NotStarted string `yaml:"not_started"`
		Starting   string `yaml:"starting"`
	}

	FaviconPath string `yaml:"favicon"`
	FaviconB64  string
}

func LoadConfig(path string) (conf Config, err error) {
	yamlFile, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("couldn't load config: %v", err)
	}

	err = yaml.Unmarshal(yamlFile, &conf)
	if err != nil {
		return Config{}, fmt.Errorf("couldn't parse config: %v", err)
	}

	if conf.FaviconPath != "" {
		conf.FaviconB64, err = encodeFavicon(conf.FaviconPath)
		if err != nil {
			return Config{}, fmt.Errorf("couldn't encode favicon: %v", err)
		}
	}

	return conf, nil
}

func encodeFavicon(path string) (result string, err error) {
	faviconRAW, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	return "data:image/png;base64," + base64.StdEncoding.EncodeToString(faviconRAW), nil
}
