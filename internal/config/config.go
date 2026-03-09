package config

import "time"

type AgentConfig struct {
	ServerAddress  string
	PollInterval   time.Duration
	ReportInterval time.Duration
}

type ServerConfig struct {
	ServerAddress string
}
