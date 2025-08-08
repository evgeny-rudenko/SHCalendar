package main

import (
	"database/sql"
	"fmt"
	"log"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

var (
	db     *sql.DB
	dateRe = dateRegexp
)

const dbFile = "calendar.db"

func openDB(path string) (*sql.DB, error) {
	if err := ensureDir(path); err != nil {
		return nil, err
	}
	// Enable busy timeout and WAL via pragmas
	dsn := fmt.Sprintf("file:%s?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)&_pragma=synchronous(NORMAL)", path)
	d, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	// Conservative settings for SQLite in a small service
	d.SetMaxOpenConns(4)
	d.SetMaxIdleConns(4)
	d.SetConnMaxIdleTime(5 * time.Minute)
	d.SetConnMaxLifetime(30 * time.Minute)

	if err := d.Ping(); err != nil {
		return nil, err
	}
	if err := initSchema(d); err != nil {
		return nil, err
	}
	return d, nil
}

func ensureDir(path string) error {
	// Create parent dir if needed
	dir := filepath.Dir(filepath.Join(".", path))
	return mkdirAll(dir)
}

func mkdirAll(dir string) error {
	// moved here to avoid importing os in multiple files explicitly
	return mkdirAllImpl(dir)
}

// initSchema ensures schema exists (preserves data) using integer habit ids; date stored as INTEGER (unix epoch seconds)
func initSchema(d *sql.DB) error {
	_, err := d.Exec(`CREATE TABLE IF NOT EXISTS marks (
		habit INTEGER NOT NULL,
		date  INTEGER NOT NULL,
		PRIMARY KEY (habit, date)
	)`)
	if err != nil {
		return err
	}
	return nil
}

// small helper to log fatal DB issues centrally
func mustOpenDB(path string) *sql.DB {
	d, err := openDB(path)
	if err != nil {
		log.Fatalf("DB open error: %v", err)
	}
	return d
}
