package store

import (
	"database/sql"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type PostgresStorage struct {
	db *sql.DB
}

func NewPostgresStorage(databaseDSN string) (*PostgresStorage, error) {
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
	/*_, err := s.db.Exec(`
		CREATE TABLE IF NOT EXISTS metrics (
			id TEXT PRIMARY KEY,
			type TEXT NOT NULL,
			delta BIGINT,
			value DOUBLE PRECISION
		)
	`)*/
	return nil
}

func (s *PostgresStorage) SetGauge(name string, value float64) error {

	/*_, err := s.db.Exec(`
		INSERT INTO metrics (id, type, value, delta)
		VALUES ($1, $2, $3, NULL)
		ON CONFLICT (id) DO UPDATE
		SET type = EXCLUDED.type,
		    value = EXCLUDED.value,
		    delta = NULL
	`, name, model.Gauge, value)
	return err*/
	return nil
}

func (s *PostgresStorage) AddCounter(name string, delta int64) (int64, error) {

	/*var result int64
	err := s.db.QueryRow(`
		INSERT INTO metrics (id, type, delta, value)
		VALUES ($1, $2, $3, NULL)
		ON CONFLICT (id) DO UPDATE
		SET type = EXCLUDED.type,
		    delta = metrics.delta + EXCLUDED.delta,
		    value = NULL
		RETURNING delta
	`, name, model.Counter, delta).Scan(&result)
	return result, err*/
	return 0, nil
}

func (s *PostgresStorage) GetGauge(name string) (float64, bool, error) {

	/*var value float64
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
	*/
	return 0, false, nil
}

func (s *PostgresStorage) GetCounter(name string) (int64, bool, error) {

	/*var delta int64
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
	*/
	return 0, false, nil
}

func (s *PostgresStorage) GetAllGauges() (map[string]float64, error) {

	/*rows, err := s.db.Query(`
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
	return result, nil*/

	return nil, nil
}

func (s *PostgresStorage) GetAllCounters() (map[string]int64, error) {

	/*rows, err := s.db.Query(`
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
	return result, nil*/

	return nil, nil
}
