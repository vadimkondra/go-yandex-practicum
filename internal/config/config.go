package config

type AgentConfig struct {
	ServerAddress  string
	PollInterval   int
	ReportInterval int
	Key            string
	RateLimit      int
}

type ServerConfig struct {
	ServerAddress string
	StoreInterval int
	FileStorePath string
	Restore       bool
	DatabaseDSN   string
	Key           string
}
