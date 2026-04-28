package config

type AgentConfig struct {
	ServerAddress  string
	PollInterval   int
	ReportInterval int
}

type ServerConfig struct {
	ServerAddress string
	StoreInterval int
	FileStorePath string
	Restore       bool
	DatabaseDsn   string
}
