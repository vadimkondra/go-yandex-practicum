package store

import "go-yandex-practicum/internal/config"

func NewStorage(cfg config.ServerConfig) (Storage, error) {

	if cfg.DatabaseDsn != "" {
		return NewPostgresStorage(cfg.DatabaseDsn)
	}
	if cfg.FileStorePath != "" {
		return NewFileStorage(cfg.FileStorePath, cfg.StoreInterval, cfg.Restore)
	}
	return NewMemoryStorage(), nil
}
