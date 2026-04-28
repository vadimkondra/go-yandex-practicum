package store

import (
	"database/sql"
	"embed"
	"fmt"
	"log"

	"go-yandex-practicum/internal/model"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

//go:embed migrations/*.sql
var embedMigrations embed.FS

type PostgresStorage struct {
	db *sql.DB
}

func NewPostgresStorage(databaseDSN string) (*PostgresStorage, error) {
	log.Printf("DATABASE_DSN = %q", databaseDSN)
	db, err := sql.Open("pgx", databaseDSN)
	if err != nil {
		return nil, err
	}
	storage := &PostgresStorage{db: db}
	if err := storage.Ping(); err != nil {
		db.Close()
		return nil, err
	}
	if err := storage.createTables(); err != nil {
		db.Close()
		return nil, err
	}
	return storage, nil
}

func (s *PostgresStorage) Ping() error {
	return s.db.Ping()
}

func (s *PostgresStorage) Close() error {
	return s.db.Close()
}

func (s *PostgresStorage) createTables() error {
	goose.SetBaseFS(embedMigrations)

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("set goose dialect: %w", err)
	}

	if err := goose.Up(s.db, "migrations"); err != nil {
		return fmt.Errorf("run postgres migrations: %w", err)
	}

	return nil
}

func (s *PostgresStorage) SetGauge(name string, value float64) error {
	_, err := s.db.Exec(`
		INSERT INTO metrics (id, type, value, delta)
		VALUES ($1, $2, $3, NULL)
		ON CONFLICT (id) DO UPDATE
		SET type = EXCLUDED.type,
		    value = EXCLUDED.value,
		    delta = NULL
	`, name, models.Gauge, value)

	return err
}

func (s *PostgresStorage) AddCounter(name string, delta int64) (int64, error) {
	var result int64

	err := s.db.QueryRow(`
		INSERT INTO metrics (id, type, delta, value)
		VALUES ($1, $2, $3, NULL)
		ON CONFLICT (id) DO UPDATE
		SET type = EXCLUDED.type,
		    delta = COALESCE(metrics.delta, 0) + EXCLUDED.delta,
		    value = NULL
		RETURNING delta
	`, name, models.Counter, delta).Scan(&result)

	return result, err
}

func (s *PostgresStorage) GetGauge(name string) (float64, bool, error) {
	var value float64

	err := s.db.QueryRow(`
		SELECT value
		FROM metrics
		WHERE id = $1 AND type = $2
	`, name, models.Gauge).Scan(&value)
	if err == sql.ErrNoRows {
		return 0, false, nil
	}
	if err != nil {
		return 0, false, err
	}

	return value, true, nil
}

func (s *PostgresStorage) GetCounter(name string) (int64, bool, error) {
	var delta int64

	err := s.db.QueryRow(`
		SELECT delta
		FROM metrics
		WHERE id = $1 AND type = $2
	`, name, models.Counter).Scan(&delta)
	if err == sql.ErrNoRows {
		return 0, false, nil
	}
	if err != nil {
		return 0, false, err
	}

	return delta, true, nil
}

func (s *PostgresStorage) GetAllGauges() (map[string]float64, error) {
	rows, err := s.db.Query(`
		SELECT id, value
		FROM metrics
		WHERE type = $1
	`, models.Gauge)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string]float64)
	for rows.Next() {
		var name string
		var value float64

		if err := rows.Scan(&name, &value); err != nil {
			return nil, err
		}

		result[name] = value
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

func (s *PostgresStorage) GetAllCounters() (map[string]int64, error) {
	rows, err := s.db.Query(`
		SELECT id, delta
		FROM metrics
		WHERE type = $1
	`, models.Counter)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string]int64)
	for rows.Next() {
		var name string
		var delta int64

		if err := rows.Scan(&name, &delta); err != nil {
			return nil, err
		}

		result[name] = delta
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}
