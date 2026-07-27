package cache

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
)

type Cache struct {
	db *sql.DB
}

func Open(path string) (*Cache, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("cache: open %q: %w", path, err)
	}

	db.SetMaxOpenConns(1)

	if _, err := db.Exec(schema); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("cache: create schema: %w", err)
	}

	return &Cache{db: db}, nil
}

func (c *Cache) Close() error {
	return c.db.Close()
}

func (c *Cache) Get(url string, parserVersion int) ([]byte, bool, error) {
	var body []byte
	var expiresAt sql.NullInt64

	err := c.db.QueryRow(
		`SELECT body, expires_at FROM responses
		 WHERE url = ? AND parser_version = ?`,
		url, parserVersion,
	).Scan(&body, &expiresAt)

	switch {
	case errors.Is(err, sql.ErrNoRows):
		return nil, false, nil
	case err != nil:
		return nil, false, fmt.Errorf("cache: get: %w", err)
	}

	if expiresAt.Valid && time.Now().Unix() >= expiresAt.Int64 {
		return nil, false, nil
	}

	return body, true, nil
}

func (c *Cache) Set(url string, parserVersion int, body []byte, expiresAt time.Time) error {
	var expires sql.NullInt64
	if !expiresAt.IsZero() {
		expires = sql.NullInt64{Int64: expiresAt.Unix(), Valid: true}
	}

	_, err := c.db.Exec(
		`INSERT INTO responses (url, parser_version, fetched_at, expires_at, body)
		 VALUES (?, ?, ?, ?, ?)
		 ON CONFLICT (url, parser_version) DO UPDATE SET
		   fetched_at  = excluded.fetched_at,
		   expires_at  = excluded.expires_at,
		   body        = excluded.body`,
		url, parserVersion, time.Now().Unix(), expires, body,
	)
	if err != nil {
		return fmt.Errorf("cache: set: %w", err)
	}
	return nil
}

const schema = `
CREATE TABLE IF NOT EXISTS responses (
	url            TEXT    NOT NULL,
	parser_version INTEGER NOT NULL,
	fetched_at     INTEGER NOT NULL,
	expires_at     INTEGER,
	body           BLOB    NOT NULL,
	PRIMARY KEY (url, parser_version)
);
`
