package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

var DB *sql.DB

func Init(dataDir string) error {
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return fmt.Errorf("create data dir: %w", err)
	}

	dbPath := filepath.Join(dataDir, "link-anime.db")
	var err error
	DB, err = sql.Open("sqlite", dbPath+"?_journal_mode=WAL&_busy_timeout=5000")
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}

	DB.SetMaxOpenConns(1)

	if err := migrate(); err != nil {
		return fmt.Errorf("migrate: %w", err)
	}

	log.Printf("Database initialized at %s", dbPath)
	return nil
}

func Close() {
	if DB != nil {
		DB.Close()
	}
}

func migrate() error {
	migrations := []string{
		`CREATE TABLE IF NOT EXISTS settings (
			key   TEXT PRIMARY KEY,
			value TEXT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS history (
			id         INTEGER PRIMARY KEY AUTOINCREMENT,
			timestamp  DATETIME DEFAULT CURRENT_TIMESTAMP,
			media_type TEXT NOT NULL,
			show_name  TEXT NOT NULL,
			season     INTEGER,
			file_count INTEGER NOT NULL,
			total_size INTEGER NOT NULL DEFAULT 0,
			dest_path  TEXT NOT NULL,
			source     TEXT NOT NULL DEFAULT ''
		)`,
		`CREATE TABLE IF NOT EXISTS linked_files (
			id          INTEGER PRIMARY KEY AUTOINCREMENT,
			history_id  INTEGER REFERENCES history(id) ON DELETE CASCADE,
			file_path   TEXT NOT NULL,
			source_path TEXT NOT NULL DEFAULT ''
		)`,
		`CREATE TABLE IF NOT EXISTS rss_rules (
			id           INTEGER PRIMARY KEY AUTOINCREMENT,
			name         TEXT NOT NULL,
			query        TEXT NOT NULL,
			show_name    TEXT NOT NULL,
			season       INTEGER DEFAULT 1,
			media_type   TEXT DEFAULT 'series',
			min_seeders  INTEGER DEFAULT 1,
			resolution   TEXT DEFAULT '',
			enabled      BOOLEAN DEFAULT 1,
			last_check   DATETIME,
			created_at   DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS rss_matches (
			id       INTEGER PRIMARY KEY AUTOINCREMENT,
			rule_id  INTEGER REFERENCES rss_rules(id) ON DELETE CASCADE,
			title    TEXT NOT NULL,
			hash     TEXT UNIQUE NOT NULL,
			matched  DATETIME DEFAULT CURRENT_TIMESTAMP,
			status   TEXT DEFAULT 'downloaded'
		)`,
	}

	for _, m := range migrations {
		if _, err := DB.Exec(m); err != nil {
			return fmt.Errorf("migration failed: %w\nSQL: %s", err, m)
		}
	}

	// Column additions (ALTER TABLE doesn't support IF NOT EXISTS in SQLite,
	// so we ignore "duplicate column" errors)
	alterMigrations := []string{
		`ALTER TABLE rss_rules ADD COLUMN filter TEXT DEFAULT 'noremakes'`,
		`ALTER TABLE rss_rules ADD COLUMN category TEXT DEFAULT '1_2'`,
		`ALTER TABLE rss_rules ADD COLUMN groups TEXT DEFAULT ''`,
		`ALTER TABLE rss_rules ADD COLUMN auto_link BOOLEAN DEFAULT 1`,
		`ALTER TABLE rss_matches ADD COLUMN torrent_name TEXT DEFAULT ''`,
	}

	for _, m := range alterMigrations {
		DB.Exec(m) // ignore "duplicate column" errors
	}

	return nil
}

// GetSetting retrieves a setting by key.
func GetSetting(key string) (string, error) {
	var value string
	err := DB.QueryRow("SELECT value FROM settings WHERE key = ?", key).Scan(&value)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return value, err
}

// SetSetting upserts a setting.
func SetSetting(key, value string) error {
	_, err := DB.Exec(
		"INSERT INTO settings (key, value) VALUES (?, ?) ON CONFLICT(key) DO UPDATE SET value = excluded.value",
		key, value,
	)
	return err
}
