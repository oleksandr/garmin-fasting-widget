package main

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
)

// Store defines persistent storage operations for fasting records.
type Store interface {
	Close() error
	UpsertFasts(ctx context.Context, fasts []FastRecord, device string, defaultSyncedAt int64) (int, error)
	GetHistory(ctx context.Context, limit int) ([]FastRecord, error)
	GetStats(ctx context.Context) (StatsResponse, error)
	Count(ctx context.Context) (int, error)
}

// DB implements Store using SQLite.
type DB struct {
	db *sql.DB
}

// NewDB opens or creates a SQLite database at the given path/DSN.
func NewDB(dsn string) (*DB, error) {
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("opening sqlite database: %w", err)
	}

	// Single connection ensures clean serial writes for SQLite
	db.SetMaxOpenConns(1)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("pinging sqlite database: %w", err)
	}

	d := &DB{db: db}
	if err := d.initSchema(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("initializing database schema: %w", err)
	}

	return d, nil
}

func (d *DB) initSchema(ctx context.Context) error {
	const schema = `
	PRAGMA busy_timeout = 5000;

	CREATE TABLE IF NOT EXISTS fasts (
		id INTEGER PRIMARY KEY,
		start_epoch INTEGER NOT NULL,
		end_epoch INTEGER NOT NULL,
		duration_hours REAL NOT NULL,
		target_hours INTEGER NOT NULL,
		state TEXT NOT NULL,
		device TEXT NOT NULL DEFAULT '',
		synced_at INTEGER NOT NULL
	);

	CREATE INDEX IF NOT EXISTS idx_fasts_start_epoch ON fasts(start_epoch DESC);
	`
	_, err := d.db.ExecContext(ctx, schema)
	return err
}

// Close closes the underlying SQLite database connection.
func (d *DB) Close() error {
	return d.db.Close()
}

// UpsertFasts batch-inserts or updates fasting records in a single transaction.
func (d *DB) UpsertFasts(ctx context.Context, fasts []FastRecord, device string, defaultSyncedAt int64) (int, error) {
	if len(fasts) == 0 {
		return 0, nil
	}

	if defaultSyncedAt <= 0 {
		defaultSyncedAt = time.Now().Unix()
	}

	tx, err := d.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("beginning transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	const query = `
	INSERT INTO fasts (id, start_epoch, end_epoch, duration_hours, target_hours, state, device, synced_at)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	ON CONFLICT(id) DO UPDATE SET
		start_epoch = excluded.start_epoch,
		end_epoch = excluded.end_epoch,
		duration_hours = excluded.duration_hours,
		target_hours = excluded.target_hours,
		state = excluded.state,
		device = CASE WHEN excluded.device != '' THEN excluded.device ELSE fasts.device END,
		synced_at = excluded.synced_at;
	`

	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return 0, fmt.Errorf("preparing upsert statement: %w", err)
	}
	defer stmt.Close()

	inserted := 0
	for _, f := range fasts {
		recordID := f.ID
		if recordID <= 0 {
			recordID = f.Start
		}
		if recordID <= 0 {
			continue
		}

		duration := f.GetDuration()
		target := f.GetTarget()
		recDevice := f.Device
		if recDevice == "" {
			recDevice = device
		}

		syncedAt := f.SyncedAt
		if syncedAt <= 0 {
			syncedAt = defaultSyncedAt
		}

		_, err := stmt.ExecContext(ctx,
			recordID,
			f.Start,
			f.End,
			duration,
			target,
			f.State,
			recDevice,
			syncedAt,
		)
		if err != nil {
			return 0, fmt.Errorf("executing upsert for record %d: %w", recordID, err)
		}
		inserted++
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("committing transaction: %w", err)
	}

	return inserted, nil
}

// GetHistory returns up to limit fasting records sorted from newest to oldest.
func (d *DB) GetHistory(ctx context.Context, limit int) ([]FastRecord, error) {
	limit = min(max(limit, 1), 100)

	const query = `
	SELECT id, start_epoch, end_epoch, duration_hours, target_hours, state, device, synced_at
	FROM fasts
	ORDER BY start_epoch DESC
	LIMIT ?;
	`

	rows, err := d.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("querying history: %w", err)
	}
	defer rows.Close()

	var records []FastRecord
	for rows.Next() {
		var r FastRecord
		if err := rows.Scan(
			&r.ID,
			&r.Start,
			&r.End,
			&r.DurationHours,
			&r.TargetHours,
			&r.State,
			&r.Device,
			&r.SyncedAt,
		); err != nil {
			return nil, fmt.Errorf("scanning fast record: %w", err)
		}
		r.Duration = r.DurationHours
		r.Target = r.TargetHours
		r.Synced = true
		records = append(records, r)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating history rows: %w", err)
	}

	return records, nil
}

// GetStats returns aggregated fasting metrics.
func (d *DB) GetStats(ctx context.Context) (StatsResponse, error) {
	const query = `
	SELECT
		COUNT(*),
		COALESCE(SUM(duration_hours), 0.0),
		COALESCE(AVG(duration_hours), 0.0),
		COALESCE(MAX(duration_hours), 0.0)
	FROM fasts;
	`

	var stats StatsResponse
	stats.Status = "success"

	row := d.db.QueryRowContext(ctx, query)
	if err := row.Scan(&stats.TotalFasts, &stats.TotalHours, &stats.AverageHours, &stats.LongestHours); err != nil {
		return StatsResponse{}, fmt.Errorf("scanning stats: %w", err)
	}

	return stats, nil
}

// Count returns the total number of records currently stored.
func (d *DB) Count(ctx context.Context) (int, error) {
	var count int
	err := d.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM fasts;").Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("counting records: %w", err)
	}
	return count, nil
}
