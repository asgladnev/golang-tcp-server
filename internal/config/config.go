package config

import "time"

type ServerConfig struct {
	Host        string        `yaml:"host"`
	Port        int           `yaml:"port"`
	ChunkSize   int           `yaml:"chunk_size"`
	ReadTimeout time.Duration `yaml:"read_timeout"`
	KeepAlive   time.Duration `yaml:"keep_alive"`
	MaxBuffer   uint64        `yaml:"max_buffer"`
}
