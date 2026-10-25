package main

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"
	"time"
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

func TestDB_GetDashboardData(t *testing.T) {
	ctx := t.Context()
	db := newTestDB(t)

	now := time.Now()
	// Insert completed fast and active fast
	fasts := []FastRecord{
		{
			ID:            101,
			Start:         now.Add(-20 * time.Hour).Unix(),
			End:           now.Add(-4 * time.Hour).Unix(),
			DurationHours: 16.0,
			TargetHours:   16,
			State:         "COMPLETED",
		},
		{
			ID:          102,
			Start:       now.Add(-2 * time.Hour).Unix(),
			End:         0,
			TargetHours: 16,
			State:       "ACTIVE",
		},
	}

	if _, err := db.UpsertFasts(ctx, fasts, "TestWatch", now.Unix()); err != nil {
		t.Fatalf("upsert failed: %v", err)
	}

	data, err := db.GetDashboardData(ctx)
	if err != nil {
		t.Fatalf("GetDashboardData error: %v", err)
	}

	if data.TotalFasts != 2 {
		t.Errorf("expected 2 total fasts, got %d", data.TotalFasts)
	}
	if data.ActiveFast == nil {
		t.Fatalf("expected active fast to be detected")
	}
	if data.ActiveFast.ID != 102 {
		t.Errorf("expected active fast ID 102, got %d", data.ActiveFast.ID)
	}
	if len(data.RecentFasts) != 2 {
		t.Errorf("expected 2 recent fasts in dashboard, got %d", len(data.RecentFasts))
	}
	if len(data.WeeklyOverview) != 7 {
		t.Errorf("expected 7 days in weekly overview, got %d", len(data.WeeklyOverview))
	}
}

func TestCalculateStreakAndOverview(t *testing.T) {
	now := time.Date(2026, 10, 25, 12, 0, 0, 0, time.UTC)

	// Streak with today and yesterday completed
	records := []FastRecord{
		{
			ID:            1,
			Start:         now.Unix(),
			DurationHours: 16.0,
			TargetHours:   16,
			State:         "COMPLETED",
		},
		{
			ID:            2,
			Start:         now.AddDate(0, 0, -1).Unix(),
			DurationHours: 16.5,
			TargetHours:   16,
			State:         "COMPLETED",
		},
		{
			ID:            3,
			Start:         now.AddDate(0, 0, -2).Unix(),
			DurationHours: 18.0,
			TargetHours:   16,
			State:         "COMPLETED",
		},
	}

	streak := calculateStreak(records, now)
	if streak != 3 {
		t.Errorf("expected streak 3, got %d", streak)
	}

	bars := calculateWeeklyOverview(records, now)
	if len(bars) != 7 {
		t.Fatalf("expected 7 bars, got %d", len(bars))
	}
	// Last bar is today (index 6)
	if !bars[6].Achieved || bars[6].Hours < 16.0 {
		t.Errorf("expected today to be achieved, got %+v", bars[6])
	}
}

func TestCalculateHeatmap(t *testing.T) {
	now := time.Date(2026, 10, 25, 14, 30, 0, 0, time.UTC)

	records := []FastRecord{
		{
			ID:            1,
			Start:         now.Unix(), // Today: 17h -> Goal Met (Level 3)
			DurationHours: 17.0,
			TargetHours:   16,
			State:         "COMPLETED",
		},
		{
			ID:            2,
			Start:         now.AddDate(0, 0, -1).Unix(), // Yesterday: 21h -> Extended (Level 4)
			DurationHours: 21.0,
			TargetHours:   16,
			State:         "COMPLETED",
		},
		{
			ID:            3,
			Start:         now.AddDate(0, 0, -2).Unix(), // 2 days ago: 14h -> Level 2
			DurationHours: 14.0,
			TargetHours:   16,
			State:         "COMPLETED",
		},
		{
			ID:            4,
			Start:         now.AddDate(0, 0, -3).Unix(), // 3 days ago: 8h -> Level 1
			DurationHours: 8.0,
			TargetHours:   16,
			State:         "COMPLETED",
		},
		{
			ID:            5,
			Start:         now.AddDate(0, 0, -100).Unix(), // 100 days ago (outside 90-day window)
			DurationHours: 20.0,
			TargetHours:   16,
			State:         "COMPLETED",
		},
	}

	heatmap := calculateHeatmap(records, now)

	if len(heatmap.Weeks) < 13 || len(heatmap.Weeks) > 15 {
		t.Errorf("expected 13-14 weeks for 90 days, got %d", len(heatmap.Weeks))
	}

	if heatmap.ActiveDays != 4 {
		t.Errorf("expected 4 active days within 90-day window, got %d", heatmap.ActiveDays)
	}

	if heatmap.Total90dFasts != 4 {
		t.Errorf("expected 4 fasts within 90-day window, got %d", heatmap.Total90dFasts)
	}

	expectedHours := 17.0 + 21.0 + 14.0 + 8.0 // 60.0h
	if heatmap.Total90dHours != fmt.Sprintf("%.1f", expectedHours) {
		t.Errorf("expected Total90dHours %.1f, got %s", expectedHours, heatmap.Total90dHours)
	}

	// 2 out of 4 days met the 16h target (17h and 21h) -> 50%
	if heatmap.SuccessRate != "50%" {
		t.Errorf("expected SuccessRate 50%%, got %s", heatmap.SuccessRate)
	}

	// Check today cell in the last week
	lastWeek := heatmap.Weeks[len(heatmap.Weeks)-1]
	var foundToday bool
	for _, day := range lastWeek.Days {
		if day.IsToday {
			foundToday = true
			if day.Level != 3 {
				t.Errorf("expected today level 3 (17h), got %d", day.Level)
			}
			if !day.Achieved {
				t.Errorf("expected today to have Achieved = true")
			}
			if !strings.Contains(day.Tooltip, "Goal Met!") {
				t.Errorf("expected tooltip to indicate goal met, got: %s", day.Tooltip)
			}
		}
	}
	if !foundToday {
		t.Errorf("today cell not found in last week")
	}

	// Verify month labels are non-empty for start and transitions
	var monthLabels []string
	for _, w := range heatmap.Weeks {
		if w.MonthLabel != "" {
			monthLabels = append(monthLabels, w.MonthLabel)
		}
	}
	if len(monthLabels) < 3 {
		t.Errorf("expected at least 3 month labels across 90 days, got %v", monthLabels)
	}
}
