package main

import (
	"cmp"
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

// Store defines persistent storage operations for fasting records.
type Store interface {
	Close() error
	UpsertFasts(ctx context.Context, fasts []FastRecord, device string, defaultSyncedAt int64) (int, error)
	GetHistory(ctx context.Context, limit int) ([]FastRecord, error)
	GetStats(ctx context.Context) (StatsResponse, error)
	GetDashboardData(ctx context.Context) (DashboardData, error)
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

	if stats.TotalFasts > 0 {
		var completedCount int
		err := d.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM fasts WHERE duration_hours >= target_hours;").Scan(&completedCount)
		if err == nil {
			stats.SuccessRatePct = (float64(completedCount) / float64(stats.TotalFasts)) * 100.0
		}
	}

	// Fetch recent records to compute streak, weekly breakdown, and 90-day heatmap
	recent, err := d.GetHistory(ctx, 300)
	if err == nil {
		now := time.Now()
		stats.CurrentStreak = calculateStreak(recent, now)
		stats.WeeklyOverview = calculateWeeklyOverview(recent, now)
		heatmap := calculateHeatmap(recent, now)
		stats.Heatmap = &heatmap
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

// GetDashboardData aggregates metrics and formats history records for the web dashboard.
func (d *DB) GetDashboardData(ctx context.Context) (DashboardData, error) {
	stats, err := d.GetStats(ctx)
	if err != nil {
		return DashboardData{}, err
	}

	history, err := d.GetHistory(ctx, 30)
	if err != nil {
		return DashboardData{}, err
	}

	var heatmap HeatmapData
	if stats.Heatmap != nil {
		heatmap = *stats.Heatmap
	} else {
		heatmap = calculateHeatmap(history, time.Now())
	}

	data := DashboardData{
		TotalFasts:     stats.TotalFasts,
		TotalHours:     fmt.Sprintf("%.1f", stats.TotalHours),
		AverageHours:   fmt.Sprintf("%.1f", stats.AverageHours),
		LongestHours:   fmt.Sprintf("%.1f", stats.LongestHours),
		Streak:         stats.CurrentStreak,
		SuccessRate:    fmt.Sprintf("%.0f%%", stats.SuccessRatePct),
		WeeklyOverview: stats.WeeklyOverview,
		Heatmap:        heatmap,
		ServerTime:     time.Now().Format("15:04:05 MST"),
	}

	views := make([]FastRecordView, 0, len(history))
	nowUnix := time.Now().Unix()

	for _, r := range history {
		view := formatRecordView(r)
		views = append(views, view)

		// Detect active fast: explicit FASTING/ACTIVE state or unset end time with past start
		isFastingState := strings.EqualFold(r.State, "FASTING") || strings.EqualFold(r.State, "ACTIVE")
		isActiveTime := (r.End == 0 || r.End > nowUnix) && r.Start <= nowUnix && !strings.EqualFold(r.State, "COMPLETED")

		if data.ActiveFast == nil && (isFastingState || isActiveTime) {
			activeCopy := view
			data.ActiveFast = &activeCopy
		}
	}
	data.RecentFasts = views

	return data, nil
}

func formatRecordView(r FastRecord) FastRecordView {
	durationH := r.GetDuration()
	targetH := r.GetTarget()

	startTime := time.Unix(r.Start, 0)
	var endTime time.Time
	if r.End > 0 {
		endTime = time.Unix(r.End, 0)
	}

	isFastingState := strings.EqualFold(r.State, "FASTING") || strings.EqualFold(r.State, "ACTIVE")

	// If currently active and duration not finalized, compute elapsed duration
	if durationH == 0 && (isFastingState || r.End == 0) && r.Start > 0 {
		elapsed := time.Since(startTime).Hours()
		if elapsed > 0 {
			durationH = elapsed
		}
	}

	// Format duration e.g. "16h 30m"
	totalMinutes := int(durationH * 60)
	dh := totalMinutes / 60
	dm := totalMinutes % 60
	durationStr := fmt.Sprintf("%dh %02dm", dh, dm)
	if totalMinutes == 0 && durationH > 0 {
		durationStr = fmt.Sprintf("%.1fh", durationH)
	}

	targetStr := fmt.Sprintf("%dh", targetH)

	badgeClass := "badge-completed"
	badgeText := "Completed"

	if isFastingState || r.End == 0 {
		badgeClass = "badge-active"
		badgeText = "In Progress"
	} else if durationH >= float64(targetH) && targetH > 0 {
		badgeClass = "badge-goal"
		badgeText = "Goal Met"
	}

	endFormatted := "--:--"
	endDateFormatted := "--"
	if r.End > 0 {
		endFormatted = endTime.Format("15:04")
		endDateFormatted = endTime.Format("Jan 02")
	}

	syncedAtStr := ""
	if r.SyncedAt > 0 {
		syncedAtStr = time.Unix(r.SyncedAt, 0).Format("Jan 02 15:04")
	}

	device := r.Device
	if device == "" {
		device = "Garmin Watch"
	}

	return FastRecordView{
		ID:               r.ID,
		StartDate:        startTime.Format("Mon, Jan 02"),
		StartTime:        startTime.Format("15:04"),
		EndDate:          endDateFormatted,
		EndTime:          endFormatted,
		DurationStr:      durationStr,
		TargetStr:        targetStr,
		DurationHours:    durationH,
		TargetHours:      targetH,
		State:            r.State,
		StatusBadgeClass: badgeClass,
		StatusBadgeText:  badgeText,
		Device:           device,
		SyncedAtStr:      syncedAtStr,
	}
}

// calculateStreak returns the count of consecutive calendar days with successful fasts.
func calculateStreak(records []FastRecord, now time.Time) int {
	if len(records) == 0 {
		return 0
	}

	successfulDays := make(map[string]bool)
	for _, r := range records {
		dur := r.GetDuration()
		tgt := r.GetTarget()
		if dur >= float64(tgt) && tgt > 0 {
			// A fast counts towards the day it finished or started
			if r.End > 0 {
				endDate := time.Unix(r.End, 0).Format("2006-01-02")
				successfulDays[endDate] = true
			}
			startDate := time.Unix(r.Start, 0).Format("2006-01-02")
			successfulDays[startDate] = true
		}
	}

	todayKey := now.Format("2006-01-02")
	yesterdayKey := now.AddDate(0, 0, -1).Format("2006-01-02")

	// If neither today nor yesterday has a completed fast, streak is 0
	startOffset := 0
	if successfulDays[todayKey] {
		startOffset = 0
	} else if successfulDays[yesterdayKey] {
		startOffset = 1
	} else {
		return 0
	}

	streak := 0
	for {
		checkDay := now.AddDate(0, 0, -(startOffset + streak)).Format("2006-01-02")
		if successfulDays[checkDay] {
			streak++
		} else {
			break
		}
	}

	return streak
}

// calculateWeeklyOverview builds a 7-day breakdown for the visual overview chart.
func calculateWeeklyOverview(records []FastRecord, now time.Time) []DayBar {
	bars := make([]DayBar, 7)

	for i := range 7 {
		offset := 6 - i
		day := now.AddDate(0, 0, -offset)
		dayKey := day.Format("2006-01-02")

		var dayHours float64
		var target float64 = 16.0

		for _, r := range records {
			rDay := time.Unix(r.Start, 0).Format("2006-01-02")
			if rDay == dayKey {
				dayHours += r.GetDuration()
				if r.GetTarget() > 0 {
					target = float64(r.GetTarget())
				}
			}
		}

		achieved := dayHours >= target && dayHours > 0
		pct := int((dayHours / 24.0) * 100.0)
		if pct > 100 {
			pct = 100
		}
		if dayHours > 0 && pct < 10 {
			pct = 10
		}

		bars[i] = DayBar{
			DayName:   day.Format("Mon"),
			DateStr:   day.Format("Jan 02"),
			Hours:     dayHours,
			Target:    target,
			Achieved:  achieved,
			HeightPct: pct,
		}
	}

	return bars
}

// calculateHeatmap generates a 90-day GitHub-style activity calendar grouped into weeks.
func calculateHeatmap(records []FastRecord, now time.Time) HeatmapData {
	const totalDays = 90
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	startDate := today.AddDate(0, 0, -(totalDays - 1))
	todayKey := today.Format("2006-01-02")

	dayHours := make(map[string]float64)
	dayTargets := make(map[string]float64)
	dayFastCount := make(map[string]int)

	for _, r := range records {
		t := time.Unix(r.Start, 0).In(now.Location())
		key := t.Format("2006-01-02")
		dur := r.GetDuration()
		if dur == 0 && (strings.EqualFold(r.State, "FASTING") || strings.EqualFold(r.State, "ACTIVE") || r.End == 0) {
			dur = time.Since(t).Hours()
		}
		dayHours[key] += dur
		dayFastCount[key]++
		if r.GetTarget() > 0 {
			dayTargets[key] = float64(r.GetTarget())
		}
	}

	weekdayOffset := (int(startDate.Weekday()) + 6) % 7 // Mon=0, ..., Sun=6
	firstMonday := startDate.AddDate(0, 0, -weekdayOffset)

	var weeks []HeatmapWeek
	var totalHours float64
	var totalFasts int
	var activeDays int
	var goalMetDays int

	lastMonth := ""
	currentMonday := firstMonday

	for currentMonday.Before(today.AddDate(0, 0, 1)) {
		var week HeatmapWeek
		var weekMonth string

		for dayIdx := range 7 {
			d := currentMonday.AddDate(0, 0, dayIdx)
			dKey := d.Format("2006-01-02")

			if d.Before(startDate) || d.After(today) {
				week.Days[dayIdx] = HeatmapDay{
					IsEmpty: true,
					Level:   0,
				}
				continue
			}

			mName := d.Format("Jan")
			if mName != lastMonth && weekMonth == "" {
				weekMonth = mName
				lastMonth = mName
			}

			hours := dayHours[dKey]
			target := cmp.Or(dayTargets[dKey], 16.0)
			achieved := hours >= target && hours > 0

			totalHours += hours
			totalFasts += dayFastCount[dKey]
			if hours > 0 {
				activeDays++
			}
			if achieved {
				goalMetDays++
			}

			level := 0
			switch {
			case hours <= 0:
				level = 0
			case hours < 12.0:
				level = 1
			case hours < 16.0:
				level = 2
			case hours < 20.0:
				level = 3
			default:
				level = 4
			}

			var tooltip string
			if hours <= 0 {
				tooltip = fmt.Sprintf("%s: No fast recorded", d.Format("Mon, Jan 02"))
			} else if achieved {
				tooltip = fmt.Sprintf("%s: %.1fh fasting (Goal Met! ⭐)", d.Format("Mon, Jan 02"), hours)
			} else {
				tooltip = fmt.Sprintf("%s: %.1fh fasting (Target: %.0fh)", d.Format("Mon, Jan 02"), hours, target)
			}

			week.Days[dayIdx] = HeatmapDay{
				DateKey:     dKey,
				DisplayDate: d.Format("Mon, Jan 02"),
				DayName:     d.Format("Mon"),
				Hours:       hours,
				Target:      target,
				Level:       level,
				Achieved:    achieved,
				Tooltip:     tooltip,
				IsToday:     (dKey == todayKey),
				IsEmpty:     false,
			}
		}

		week.MonthLabel = weekMonth
		weeks = append(weeks, week)
		currentMonday = currentMonday.AddDate(0, 0, 7)
	}

	successRateStr := "0%"
	if activeDays > 0 {
		successRateStr = fmt.Sprintf("%.0f%%", (float64(goalMetDays)/float64(activeDays))*100.0)
	}

	return HeatmapData{
		Weeks:         weeks,
		Total90dHours: fmt.Sprintf("%.1f", totalHours),
		Total90dFasts: totalFasts,
		ActiveDays:    activeDays,
		SuccessRate:   successRateStr,
	}
}
