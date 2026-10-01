package main

import (
	"path/filepath"
	"testing"
)

func newTestDB(t *testing.T) *DB {
	t.Helper()
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test_fasting.db")

	db, err := NewDB(dbPath)
	if err != nil {
		t.Fatalf("failed to create test db: %v", err)
	}
	t.Cleanup(func() {
		_ = db.Close()
	})
	return db
}

func TestDB_UpsertAndHistory(t *testing.T) {
	ctx := t.Context()
	db := newTestDB(t)

	// 1. Initial count should be 0
	count, err := db.Count(ctx)
	if err != nil {
		t.Fatalf("Count error: %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0 records, got %d", count)
	}

	// 2. Insert records
	fasts := []FastRecord{
		{
			ID:            1001,
			Start:         1727400000,
			End:           1727457600,
			DurationHours: 16.0,
			TargetHours:   16,
			State:         "COMPLETED",
		},
		{
			ID:            1002,
			Start:         1727500000,
			End:           1727561200,
			DurationHours: 17.0,
			TargetHours:   16,
			State:         "COMPLETED",
		},
	}

	inserted, err := db.UpsertFasts(ctx, fasts, "Instinct 3 AMOLED", 1727600000)
	if err != nil {
		t.Fatalf("UpsertFasts error: %v", err)
	}
	if inserted != 2 {
		t.Errorf("expected 2 inserted, got %d", inserted)
	}

	// 3. Count should now be 2
	count, err = db.Count(ctx)
	if err != nil {
		t.Fatalf("Count error: %v", err)
	}
	if count != 2 {
		t.Errorf("expected 2 records, got %d", count)
	}

	// 4. Query history (newest first)
	history, err := db.GetHistory(ctx, 10)
	if err != nil {
		t.Fatalf("GetHistory error: %v", err)
	}
	if len(history) != 2 {
		t.Fatalf("expected 2 history items, got %d", len(history))
	}
	if history[0].ID != 1002 {
		t.Errorf("expected newest record id 1002 first, got %d", history[0].ID)
	}
	if history[1].ID != 1001 {
		t.Errorf("expected second record id 1001, got %d", history[1].ID)
	}
	if history[0].Device != "Instinct 3 AMOLED" {
		t.Errorf("expected device 'Instinct 3 AMOLED', got %s", history[0].Device)
	}

	// 5. Test Upsert conflict (updating record 1002 with bonus duration)
	updatedFasts := []FastRecord{
		{
			ID:            1002,
			Start:         1727500000,
			End:           1727564800,
			DurationHours: 18.0,
			TargetHours:   16,
			State:         "COMPLETED",
		},
	}

	inserted, err = db.UpsertFasts(ctx, updatedFasts, "Instinct 3 AMOLED", 1727605000)
	if err != nil {
		t.Fatalf("UpsertFasts update error: %v", err)
	}
	if inserted != 1 {
		t.Errorf("expected 1 updated, got %d", inserted)
	}

	// Total count remains 2
	count, err = db.Count(ctx)
	if err != nil {
		t.Fatalf("Count error: %v", err)
	}
	if count != 2 {
		t.Errorf("expected count 2 after update, got %d", count)
	}

	// Verify updated duration
	history, err = db.GetHistory(ctx, 1)
	if err != nil {
		t.Fatalf("GetHistory error: %v", err)
	}
	if history[0].DurationHours != 18.0 {
		t.Errorf("expected updated duration 18.0, got %f", history[0].DurationHours)
	}
}

func TestDB_Stats(t *testing.T) {
	ctx := t.Context()
	db := newTestDB(t)

	// Empty stats
	stats, err := db.GetStats(ctx)
	if err != nil {
		t.Fatalf("GetStats error: %v", err)
	}
	if stats.TotalFasts != 0 {
		t.Errorf("expected 0 total fasts, got %d", stats.TotalFasts)
	}

	// Insert test data
	fasts := []FastRecord{
		{ID: 1, Start: 100, End: 200, DurationHours: 14.0, TargetHours: 16, State: "COMPLETED"},
		{ID: 2, Start: 300, End: 400, DurationHours: 16.0, TargetHours: 16, State: "COMPLETED"},
		{ID: 3, Start: 500, End: 600, DurationHours: 20.0, TargetHours: 16, State: "COMPLETED"},
	}
	if _, err := db.UpsertFasts(ctx, fasts, "Watch", 700); err != nil {
		t.Fatalf("Upsert error: %v", err)
	}

	stats, err = db.GetStats(ctx)
	if err != nil {
		t.Fatalf("GetStats error: %v", err)
	}
	if stats.TotalFasts != 3 {
		t.Errorf("expected 3 total fasts, got %d", stats.TotalFasts)
	}
	if stats.TotalHours != 50.0 {
		t.Errorf("expected 50.0 total hours, got %f", stats.TotalHours)
	}
	expectedAvg := 50.0 / 3.0
	if stats.AverageHours < expectedAvg-0.01 || stats.AverageHours > expectedAvg+0.01 {
		t.Errorf("expected average hours ~%f, got %f", expectedAvg, stats.AverageHours)
	}
	if stats.LongestHours != 20.0 {
		t.Errorf("expected longest 20.0, got %f", stats.LongestHours)
	}
}
