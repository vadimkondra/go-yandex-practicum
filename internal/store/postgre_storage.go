package store

import (
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"go-yandex-practicum/internal/retry"
	"log"
	"strings"

	"go-yandex-practicum/internal/model"

	"github.com/jackc/pgx/v5/pgconn"
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
	return retry.Do(func() error {
		_, err := s.db.Exec(`
		INSERT INTO metrics (id, type, value, delta)
		VALUES ($1, $2, $3, NULL)
		ON CONFLICT (id) DO UPDATE
		SET type = EXCLUDED.type,
		    value = EXCLUDED.value,
		    delta = NULL
	`, name, model.Gauge, value)

		return err
	}, isRetriablePostgresError)
}

func (s *PostgresStorage) AddCounter(name string, delta int64) (int64, error) {
	var result int64

	err := retry.Do(func() error {
		return s.db.QueryRow(`
			INSERT INTO metrics (id, type, delta, value)
			VALUES ($1, $2, $3, NULL)
			ON CONFLICT (id) DO UPDATE
			SET type = EXCLUDED.type,
			    delta = COALESCE(metrics.delta, 0) + EXCLUDED.delta,
			    value = NULL
			RETURNING delta
		`, name, model.Counter, delta).Scan(&result)
	}, isRetriablePostgresError)
	return result, err
}

func (s *PostgresStorage) GetGauge(name string) (float64, bool, error) {
	var value float64

	err := s.db.QueryRow(`
		SELECT value
		FROM metrics
		WHERE id = $1 AND type = $2
	`, name, model.Gauge).Scan(&value)
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
	`, name, model.Counter).Scan(&delta)
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
	`, model.Gauge)
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
	`, model.Counter)
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

func (s *PostgresStorage) UpdateBatch(metrics []model.Metrics) ([]model.Metrics, error) {
	var updated []model.Metrics

	err := retry.Do(func() error {
		result, err := s.updateBatchOnce(metrics)
		if err != nil {
			return err
		}

		updated = result
		return nil
	}, isRetriablePostgresError)

	return updated, err
}

func (s *PostgresStorage) updateBatchOnce(metrics []model.Metrics) ([]model.Metrics, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return nil, err
	}

	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	updated := make([]model.Metrics, 0, len(metrics))

	for _, metric := range metrics {
		switch metric.MType {
		case model.Gauge:
			if metric.Value == nil {
				continue
			}

			var value float64

			err := tx.QueryRow(`
				INSERT INTO metrics (id, type, value, delta)
				VALUES ($1, $2, $3, NULL)
				ON CONFLICT (id) DO UPDATE
				SET type = EXCLUDED.type,
				    value = EXCLUDED.value,
				    delta = NULL
				RETURNING value
			`, metric.ID, model.Gauge, *metric.Value).Scan(&value)
			if err != nil {
				return nil, err
			}

			updated = append(updated, model.Metrics{
				ID:    metric.ID,
				MType: model.Gauge,
				Value: &value,
			})

		case model.Counter:
			if metric.Delta == nil {
				continue
			}

			var delta int64

			err := tx.QueryRow(`
				INSERT INTO metrics (id, type, delta, value)
				VALUES ($1, $2, $3, NULL)
				ON CONFLICT (id) DO UPDATE
				SET type = EXCLUDED.type,
				    delta = COALESCE(metrics.delta, 0) + EXCLUDED.delta,
				    value = NULL
				RETURNING delta
			`, metric.ID, model.Counter, *metric.Delta).Scan(&delta)
			if err != nil {
				return nil, err
			}

			updated = append(updated, model.Metrics{
				ID:    metric.ID,
				MType: model.Counter,
				Delta: &delta,
			})
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	committed = true

	return updated, nil
}

func isRetriablePostgresError(err error) bool {

	if err == nil {
		return false
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return strings.HasPrefix(pgErr.Code, "08")
	}
	return false

}
