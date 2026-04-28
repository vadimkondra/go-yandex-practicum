package store

import (
	"database/sql"
	"log"
)

var db *sql.DB

func InitDb(databaseDsn string) {
	db, err := sql.Open("pgx", databaseDsn)
	if err != nil {
		log.Fatal(err)
	}

	if err := db.Ping(); err != nil {
		log.Fatal(err)
	}
}

func Ping() bool {
	if db == nil {
		return false
	}
	if err := db.Ping(); err != nil {
		return false
	}
	return true
}

func CloseDb() {
	if db != nil {
		db.Close()
	}
}
