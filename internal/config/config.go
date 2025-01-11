package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server struct {
		Port int `yaml:"port"`
	} `yaml:"server"`
	Storage struct {
		Path string `yaml:"path"`
	} `yaml:"storage"`
	Tunnel struct {
		Driver          string `yaml:"driver"`
		CloudflareToken string `yaml:"cloudflare_token,omitempty"`
	} `yaml:"tunnel"`
	Services map[string]ServiceConfig `yaml:"services"`
}

type ServiceConfig struct {
	EventTypeSource   string `yaml:"event_type_source"`
	EventTypeLocation string `yaml:"event_type_location"`
}

func getConfigLocations(configPath string) []string {
	if configPath != "" {
		return []string{configPath}
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return []string{}
	}

	return []string{filepath.Join(home, ".config", "webhook-consumer", "config.yaml")}
}

func Load(configPath string) (*Config, error) {
	locations := getConfigLocations(configPath)

	var config *Config

	for _, loc := range locations {
		data, _ := os.ReadFile(loc)

		config = &Config{}
		if err := yaml.Unmarshal(data, config); err != nil {
			return nil, fmt.Errorf("error parsing config file %s: %w", loc, err)
		}

		break // Use the first valid config file found
	}

	// If no config file was found, create default config
	if config == nil {
		config = &Config{
			Services: make(map[string]ServiceConfig),
		}
	}

	if config.Services == nil {
		config.Services = make(map[string]ServiceConfig)
	}

	// Apply defaults
	if config.Server.Port == 0 {
		config.Server.Port = 8080
	}
	if config.Storage.Path == "" {
		config.Storage.Path = "./logs"
	}
	if config.Tunnel.Driver == "" {
		config.Tunnel.Driver = "local"
	}

	return config, nil
}
