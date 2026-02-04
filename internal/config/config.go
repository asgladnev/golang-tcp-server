package config

import "time"

type ServerConfig struct {
	Host        string
	Port        int
	ChunkSize   int
	ReadTimeout time.Duration
	KeepAlive   time.Duration
	MaxBuffer   uint64
}
