package store

import "go-yandex-practicum/internal/config"

func NewStorage(cfg config.ServerConfig) (Storage, error) {

	if cfg.DatabaseDSN != "" {
		return NewPostgresStorage(cfg.DatabaseDSN)
	}
	if cfg.FileStorePath != "" {
		return NewFileStorage(cfg.FileStorePath, cfg.StoreInterval, cfg.Restore)
	}
	return NewMemoryStorage(), nil
}
