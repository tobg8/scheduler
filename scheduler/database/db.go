package database

import (
	"database/sql"
	_ "embed"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

//go:embed table_creation.sql
var table_creation string

func InitializeDB() (*sql.DB, error) {
	dbPath := "../app.db?cache=shared"
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("could not initialize database: %w", err)
	}

	_, err = db.Exec(table_creation)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("could not create schema for table issues: %w", err)
	}

	log.Print("initializing database")

	return db, nil
}
