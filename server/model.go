package main

// FastRecord represents an individual fasting session synced from the Garmin watch.
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

// SyncRequest is the JSON payload sent by the Garmin watch sync client.
type SyncRequest struct {
	Device   string       `json:"device,omitzero"`
	SyncedAt int64        `json:"synced_at,omitzero"`
	Fasts    []FastRecord `json:"fasts"`
}

// SyncResponse is the JSON response returned to the Garmin watch.
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

// StatsResponse provides summary metrics for fasting history.
type StatsResponse struct {
	Status       string  `json:"status"`
	TotalFasts   int     `json:"total_fasts"`
	TotalHours   float64 `json:"total_hours"`
	AverageHours float64 `json:"average_hours"`
	LongestHours float64 `json:"longest_hours"`
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
