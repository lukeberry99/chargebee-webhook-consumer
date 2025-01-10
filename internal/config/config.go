package config

import (
	"fmt"
	"log"
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
}

func getConfigLocations(configPath string) []string {
	if configPath != "" {
		return []string{configPath}
	}

	locations := []string{}

	if xdgConfig := os.Getenv("XDG_CONFIG_HOME"); xdgConfig != "" {
		locations = append(locations, filepath.Join(xdgConfig, "webhook-consumer", "config.yaml"))
	} else if home, err := os.UserHomeDir(); err == nil {
		// XDG default is ~/.config
		locations = append(locations, filepath.Join(home, ".config", "webhook-consumer", "config.yaml"))
	}

	if xdgConfigDirs := os.Getenv("XDG_CONFIG_DIRS"); xdgConfigDirs != "" {
		for _, dir := range filepath.SplitList(xdgConfigDirs) {
			locations = append(locations, filepath.Join(dir, "webhook-consumer", "config.yaml"))
		}
	} else {
		// Default XDG_CONFIG_DIRS is /etc/xdg
		locations = append(locations, "/etc/xdg/webhook-consumer/config.yaml")
	}

	// /etc fallback
	locations = append(locations, "/etc/webhook-consumer/config.yaml")

	// current working directory
	locations = append(locations, "config.yaml")

	return locations
}

func Load(configPath string) (*Config, error) {
	locations := getConfigLocations(configPath)

	var lastErr error
	var config *Config

	for _, loc := range locations {
		data, err := os.ReadFile(loc)
		if err != nil {
			lastErr = err
			continue
		}

		config = &Config{}
		if err := yaml.Unmarshal(data, config); err != nil {
			return nil, fmt.Errorf("error parsing config file %s: %w", loc, err)
		}

		log.Printf("Using config file: %s", loc)
		break // Use the first valid config file found
	}

	// If no config file was found, create default config
	if config == nil {
		if lastErr != nil {
			log.Printf("No config files found, using defaults.")
		}
		config = &Config{}
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
