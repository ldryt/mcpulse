package config

import (
	"encoding/base64"
	"flag"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	ListenAddress string `yaml:"listen-address"`

	Motd        string `yaml:"motd"`
	FaviconPath string `yaml:"favicon"`
	FaviconB64  string

	LoginFlow struct {
		WelcomeMessage         string `yaml:"welcome-msg"`
		BackendNotReadyMessage string `yaml:"backend-not-ready-msg"`
	} `yaml:"login-flow"`
}

var cfg *Config

func Get() *Config {
	return cfg
}

func Load() (err error) {
	configPathPTR := flag.String("config", "./config.yml", "a path to the configuration file")
	flag.Parse()
	configPath := *configPathPTR

	yamlFile, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("error loading config: %v", err)
	}

	err = yaml.Unmarshal(yamlFile, &cfg)
	if err != nil {
		return fmt.Errorf("error parsing config: %v", err)
	}

	if cfg.FaviconPath != "" {
		cfg.FaviconB64, err = encodeFavicon(cfg.FaviconPath)
		if err != nil {
			return fmt.Errorf("error encoding favicon: %v", err)
		}
	}

	return nil
}

func encodeFavicon(path string) (result string, err error) {
	faviconRAW, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	return "data:image/png;base64," + base64.StdEncoding.EncodeToString(faviconRAW), nil
}
