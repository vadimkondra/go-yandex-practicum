package store

import (
	"database/sql"
	"log"
)

type DBStorage struct {
	db *sql.DB
}

var storage DBStorage

func InitDB(databaseDSN string) {
	db, err := sql.Open("pgx", databaseDSN)
	storage := &DBStorage{db: db}
	if err != nil {
		log.Fatal(err)
	}

	if err := storage.db.Ping(); err != nil {
		log.Fatal(err)
	}
}

func Ping() bool {
	if storage.db == nil {
		return false
	}
	if err := storage.db.Ping(); err != nil {
		return false
	}
	return true
}

func CloseDB() {
	if storage.db != nil {
		storage.db.Close()
	}
}
