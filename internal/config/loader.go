package config

import (
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

func LoadFromFile(path string) (*ServerConfig, error) {
	var err error
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg ServerConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	if cfg.Host == "" {
		cfg.Host = "127.0.0.1"
	}
	if cfg.Port == 0 {
		cfg.Port = 9000
	}
	if cfg.ReadTimeout == 0 {
		cfg.ReadTimeout = 60 * time.Second
	}
	if cfg.ChunkSize == 0 {
		cfg.ChunkSize = 4096
	}
	if cfg.KeepAlive == 0 {
		cfg.ReadTimeout = 60 * time.Second
	}
	return &cfg, nil
}
