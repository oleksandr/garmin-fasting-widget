package main

// FastRecord represents an individual fasting session synced from the Garmin watch or smartphone.
type FastRecord struct {
	ID            int64   `json:"id"`
	Start         int64   `json:"start"`
	End           int64   `json:"end"`
	DurationHours float64 `json:"duration_hours,omitzero"`
	Duration      float64 `json:"duration,omitzero"`
	TargetHours   int     `json:"target_hours,omitzero"`
	Target        int     `json:"target,omitzero"`
	State         string  `json:"state"`
	Synced        bool    `json:"synced,omitzero"`
	Device        string  `json:"device,omitzero"`
	SyncedAt      int64   `json:"synced_at,omitzero"`
}

// GetDuration returns the duration in hours, checking both duration_hours and duration keys.
func (f FastRecord) GetDuration() float64 {
	if f.DurationHours > 0 {
		return f.DurationHours
	}
	return f.Duration
}

// GetTarget returns the target hours, checking both target_hours and target keys.
func (f FastRecord) GetTarget() int {
	if f.TargetHours > 0 {
		return f.TargetHours
	}
	return f.Target
}

// SyncRequest is an alias for FlexibleSyncRequest for backward compatibility.
type SyncRequest = FlexibleSyncRequest

// FlexibleSyncRequest handles batch payloads, single records, or smartphone app posts.
type FlexibleSyncRequest struct {
	Device        string       `json:"device,omitzero"`
	SyncedAt      int64        `json:"synced_at,omitzero"`
	Fasts         []FastRecord `json:"fasts,omitzero"`
	ID            int64        `json:"id,omitzero"`
	Start         int64        `json:"start,omitzero"`
	End           int64        `json:"end,omitzero"`
	DurationHours float64      `json:"duration_hours,omitzero"`
	Duration      float64      `json:"duration,omitzero"`
	TargetHours   int          `json:"target_hours,omitzero"`
	Target        int          `json:"target,omitzero"`
	State         string       `json:"state,omitzero"`
}

// ExtractFasts converts the flexible payload into a uniform slice of FastRecords.
func (r FlexibleSyncRequest) ExtractFasts() []FastRecord {
	if len(r.Fasts) > 0 {
		return r.Fasts
	}
	if r.Start > 0 || r.ID > 0 {
		rec := FastRecord{
			ID:            r.ID,
			Start:         r.Start,
			End:           r.End,
			DurationHours: r.DurationHours,
			Duration:      r.Duration,
			TargetHours:   r.TargetHours,
			Target:        r.Target,
			State:         r.State,
			Device:        r.Device,
			SyncedAt:      r.SyncedAt,
		}
		if rec.ID == 0 {
			rec.ID = rec.Start
		}
		return []FastRecord{rec}
	}
	return nil
}

// SyncResponse is the JSON response returned to the Garmin watch or app client.
type SyncResponse struct {
	Status        string `json:"status"`
	ReceivedCount int    `json:"received_count"`
	Message       string `json:"message,omitzero"`
}

// HistoryResponse contains a list of recent fasting records.
type HistoryResponse struct {
	Status  string       `json:"status"`
	Count   int          `json:"count"`
	Records []FastRecord `json:"records"`
}

// DayBar represents a daily data point for the weekly visual overview.
type DayBar struct {
	DayName   string  `json:"day_name"`
	DateStr   string  `json:"date_str"`
	Hours     float64 `json:"hours"`
	Target    float64 `json:"target"`
	Achieved  bool    `json:"achieved"`
	HeightPct int     `json:"height_pct"`
}

// StatsResponse provides summary metrics for fasting history.
type StatsResponse struct {
	Status         string       `json:"status"`
	TotalFasts     int          `json:"total_fasts"`
	TotalHours     float64      `json:"total_hours"`
	AverageHours   float64      `json:"average_hours"`
	LongestHours   float64      `json:"longest_hours"`
	CurrentStreak  int          `json:"current_streak"`
	SuccessRatePct float64      `json:"success_rate_pct"`
	WeeklyOverview []DayBar     `json:"weekly_overview,omitzero"`
	Heatmap        *HeatmapData `json:"heatmap,omitzero"`
}

// HealthResponse reports service health and uptime.
type HealthResponse struct {
	Status     string `json:"status"`
	Uptime     string `json:"uptime"`
	TotalFasts int    `json:"total_fasts"`
}

// ErrorResponse represents an HTTP error response.
type ErrorResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

// HeatmapDay represents an individual day cell in the 90-day activity heatmap.
type HeatmapDay struct {
	DateKey     string  `json:"date_key"`     // "2026-10-25"
	DisplayDate string  `json:"display_date"` // "Sun, Oct 25"
	DayName     string  `json:"day_name"`     // "Sun"
	Hours       float64 `json:"hours"`
	Target      float64 `json:"target"`
	Level       int     `json:"level"`        // 0 to 4
	Achieved    bool    `json:"achieved"`
	Tooltip     string  `json:"tooltip"`
	IsToday     bool    `json:"is_today"`
	IsEmpty     bool    `json:"is_empty"`     // padding days before start or after today
}

// HeatmapWeek represents one vertical 7-day column in the calendar heatmap.
type HeatmapWeek struct {
	MonthLabel string        `json:"month_label"` // "Aug", "Sep", etc. (displayed above column)
	Days       [7]HeatmapDay `json:"days"`        // Index 0=Mon, 1=Tue, ..., 6=Sun
}

// HeatmapData groups all weeks and summary stats for the 90-day calendar.
type HeatmapData struct {
	Weeks         []HeatmapWeek `json:"weeks"`
	Total90dHours string        `json:"total_90d_hours"`
	Total90dFasts int           `json:"total_90d_fasts"`
	ActiveDays    int           `json:"active_days"`
	SuccessRate   string        `json:"success_rate"`
}

// FastRecordView formats a FastRecord for presentation on the HTML dashboard.
type FastRecordView struct {
	ID               int64
	StartDate        string
	StartTime        string
	EndDate          string
	EndTime          string
	DurationStr      string
	TargetStr        string
	DurationHours    float64
	TargetHours      int
	State            string
	StatusBadgeClass string
	StatusBadgeText  string
	Device           string
	SyncedAtStr      string
}

// DashboardData contains all data required to render the web dashboard.
type DashboardData struct {
	TotalFasts     int
	TotalHours     string
	AverageHours   string
	LongestHours   string
	Streak         int
	SuccessRate    string
	WeeklyOverview []DayBar
	Heatmap        HeatmapData
	RecentFasts    []FastRecordView
	ActiveFast     *FastRecordView
	ServerTime     string
	ServerUptime   string
}

// APISummaryResponse describes the server state when queried via GET /api.
type APISummaryResponse struct {
	Status        string        `json:"status"`
	ServerTime    string        `json:"server_time"`
	Uptime        string        `json:"uptime"`
	Stats         StatsResponse `json:"stats"`
	RecentFasts   []FastRecord  `json:"recent_fasts"`
	Documentation string        `json:"documentation"`
}
