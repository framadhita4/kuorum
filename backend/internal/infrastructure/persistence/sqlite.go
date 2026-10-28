package persistence

import (
	"context"
	"database/sql"
	_ "embed"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

//go:embed sql/schema.sql
var schemaSQL string

// NewSqliteDB opens a connection to a SQLite database and runs initial schema migrations.
func NewSqliteDB(dsn string) (*sql.DB, error) {
	if dsn == "" {
		dsn = "kuorum.db"
	}

	// Ensure directory exists if path is specified
	dir := filepath.Dir(dsn)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, fmt.Errorf("failed to create db directory %q: %w", dir, err)
		}
	}

	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite database: %w", err)
	}

	// Configure connection pool for SQLite (single-writer friendly)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(time.Hour)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping sqlite database: %w", err)
	}

	// Enable WAL mode, busy timeout, and foreign keys
	pragmas := []string{
		"PRAGMA journal_mode = WAL;",
		"PRAGMA busy_timeout = 5000;",
		"PRAGMA foreign_keys = ON;",
		"PRAGMA synchronous = NORMAL;",
	}
	for _, pragma := range pragmas {
		if _, err := db.ExecContext(ctx, pragma); err != nil {
			db.Close()
			return nil, fmt.Errorf("failed to set pragma %q: %w", pragma, err)
		}
	}

	// Run initial schema migration
	if err := runMigrations(ctx, db); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to run database migrations: %w", err)
	}

	log.Printf("[Database] SQLite initialized successfully at %q (WAL mode enabled)", dsn)
	return db, nil
}

func runMigrations(ctx context.Context, db *sql.DB) error {
	if schemaSQL == "" {
		return fmt.Errorf("embedded schema.sql is empty")
	}

	// 1. Check if legacy research_pipelines table exists as a table
	var isTable int
	checkQuery := `SELECT COUNT(*) FROM sqlite_master WHERE type = 'table' AND name = 'research_pipelines';`
	if err := db.QueryRowContext(ctx, checkQuery).Scan(&isTable); err == nil && isTable > 0 {
		// Ensure chat_histories table exists before copying
		createChatHistories := `
			CREATE TABLE IF NOT EXISTS chat_histories (
				id TEXT PRIMARY KEY NOT NULL,
				title TEXT NOT NULL,
				initial_prompt TEXT NOT NULL,
				model TEXT NOT NULL DEFAULT '',
				current_stage TEXT NOT NULL DEFAULT 'EXPLORATION',
				status TEXT NOT NULL DEFAULT 'DRAFT',
				mode TEXT NOT NULL DEFAULT 'AUTONOMOUS',
				checkpoint TEXT NOT NULL DEFAULT 'NONE',
				selected_angle_id TEXT NOT NULL DEFAULT '',
				anchor_paper_id TEXT NOT NULL DEFAULT '',
				stage_data_json TEXT NOT NULL DEFAULT '{}',
				pinned BOOLEAN NOT NULL DEFAULT 0,
				is_archived BOOLEAN NOT NULL DEFAULT 0,
				error_message TEXT NOT NULL DEFAULT '',
				created_at TIMESTAMP NOT NULL,
				updated_at TIMESTAMP NOT NULL
			);
		`
		if _, err := db.ExecContext(ctx, createChatHistories); err != nil {
			return fmt.Errorf("failed to pre-create chat_histories for migration: %w", err)
		}

		// Migrate records from research_pipelines to chat_histories
		migrateData := `
			INSERT OR IGNORE INTO chat_histories (
				id, title, initial_prompt, model, current_stage, status, mode, checkpoint,
				selected_angle_id, anchor_paper_id, stage_data_json, pinned, is_archived,
				error_message, created_at, updated_at
			)
			SELECT
				id, title, initial_prompt, '', current_stage, status, mode, checkpoint,
				selected_angle_id, anchor_paper_id, stage_data_json, 0, 0,
				error_message, created_at, updated_at
			FROM research_pipelines;
		`
		if _, err := db.ExecContext(ctx, migrateData); err != nil {
			return fmt.Errorf("failed to migrate data from research_pipelines: %w", err)
		}

		// Drop the legacy table so research_pipelines view can be created cleanly
		if _, err := db.ExecContext(ctx, `DROP TABLE research_pipelines;`); err != nil {
			return fmt.Errorf("failed to drop legacy research_pipelines table: %w", err)
		}
	}

	// 2. Execute schema definitions
	if _, err := db.ExecContext(ctx, schemaSQL); err != nil {
		return fmt.Errorf("executing schema.sql failed: %w", err)
	}

	// 2b. Add non-breaking schema columns for thinking steps and stats if migrating from older versions
	_, _ = db.ExecContext(ctx, `ALTER TABLE chat_messages ADD COLUMN thinking_steps_json TEXT NOT NULL DEFAULT '[]';`)
	_, _ = db.ExecContext(ctx, `ALTER TABLE chat_messages ADD COLUMN stats_summary_json TEXT NOT NULL DEFAULT '{}';`)
	_, _ = db.ExecContext(ctx, `ALTER TABLE chat_histories ADD COLUMN approach TEXT NOT NULL DEFAULT 'NORMAL';`)

	// 3. Seed initial user messages for any chat histories lacking messages
	seedMessages := `
		INSERT OR IGNORE INTO chat_messages (
			id, chat_id, role, content, status, stage, steps_json,
			citations_json, quote_json, checkpoint_json, suggested_follow_ups_json,
			error_message, created_at, updated_at
		)
		SELECT
			id || '-msg-user-init', id, 'user', initial_prompt, 'done', current_stage,
			'[]', '[]', '{}', '{}', '[]', '', created_at, created_at
		FROM chat_histories
		WHERE NOT EXISTS (SELECT 1 FROM chat_messages WHERE chat_messages.chat_id = chat_histories.id);
	`
	_, _ = db.ExecContext(ctx, seedMessages)

	return nil
}
