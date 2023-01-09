// Package storage implements the storage API used by the URL Shortener.
package storage

import (
	"database/sql"
	_ "embed"
	"errors"
	"strings"

	"github.com/mattn/go-sqlite3"
)

//go:embed schema.sql
var Schema string

// TestDB returns an in-memory sqlite3 database.
func TestDB() *sql.DB {
	return openDB(":memory:")
}

// ProdDB returns an on-disk database.
func ProdDB(path string) *sql.DB {
	if strings.Index(path, ":memory:") != -1 {
		panic("for in-memory databases, use TestDB")
	}
	return openDB(path)
}

func openDB(path string) *sql.DB {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		panic(err)

	}
	if _, err := db.Exec(Schema); err != nil {
		panic(err)
	}
	return db
}

// IsErrConstraint returns true if the passed err was created as the result of a constraint violation
// in the sqlite3 database.
func IsErrConstraint(err error) bool {
	var dbErr sqlite3.Error
	return errors.As(err, &dbErr) && dbErr.Code == sqlite3.ErrConstraint
}
