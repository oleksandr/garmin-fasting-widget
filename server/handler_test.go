package main

import (
	"bytes"
	"encoding/json/v2"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func setupTestServer(t *testing.T, authToken string) (*Server, *DB) {
	t.Helper()
	db := newTestDB(t)
	srv := NewServer(db, authToken)
	return srv, db
}

func TestSyncHandler_Success(t *testing.T) {
	srv, db := setupTestServer(t, "")

	payload := SyncRequest{
		Device:   "Instinct 3 AMOLED 50mm",
		SyncedAt: 1727780000,
		Fasts: []FastRecord{
			{
				ID:            501,
				Start:         1727700000,
				End:           1727757600,
				DurationHours: 16.0,
				TargetHours:   16,
				State:         "COMPLETED",
			},
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	req := httptest.NewRequest("POST", "/api/fasting/sync", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	srv.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp SyncResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Status != "success" {
		t.Errorf("expected status 'success', got %s", resp.Status)
	}
	if resp.ReceivedCount != 1 {
		t.Errorf("expected received_count 1, got %d", resp.ReceivedCount)
	}

	// Verify in database
	count, err := db.Count(t.Context())
	if err != nil {
		t.Fatalf("db count failed: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 record in db, got %d", count)
	}
}

func TestSyncHandler_Auth(t *testing.T) {
	const secret = "test-secret-token"
	srv, _ := setupTestServer(t, secret)

	payload := SyncRequest{
		Device: "Instinct 3",
		Fasts: []FastRecord{
			{ID: 1, Start: 100, End: 200, DurationHours: 16, TargetHours: 16, State: "COMPLETED"},
		},
	}
	body, _ := json.Marshal(payload)

	// 1. Missing Authorization header -> 401
	reqMissing := httptest.NewRequest("POST", "/api/fasting/sync", bytes.NewReader(body))
	wMissing := httptest.NewRecorder()
	srv.Handler().ServeHTTP(wMissing, reqMissing)
	if wMissing.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 for missing token, got %d", wMissing.Code)
	}

	// 2. Wrong token -> 401
	reqWrong := httptest.NewRequest("POST", "/api/fasting/sync", bytes.NewReader(body))
	reqWrong.Header.Set("Authorization", "Bearer invalid-token")
	wWrong := httptest.NewRecorder()
	srv.Handler().ServeHTTP(wWrong, reqWrong)
	if wWrong.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 for invalid token, got %d", wWrong.Code)
	}

	// 3. Valid token -> 200
	reqValid := httptest.NewRequest("POST", "/api/fasting/sync", bytes.NewReader(body))
	reqValid.Header.Set("Authorization", "Bearer "+secret)
	wValid := httptest.NewRecorder()
	srv.Handler().ServeHTTP(wValid, reqValid)
	if wValid.Code != http.StatusOK {
		t.Errorf("expected 200 for valid token, got %d: %s", wValid.Code, wValid.Body.String())
	}
}

func TestSyncHandler_InvalidJSON(t *testing.T) {
	srv, _ := setupTestServer(t, "")

	req := httptest.NewRequest("POST", "/api/fasting/sync", bytes.NewReader([]byte(`{"fasts": [invalid`)))
	w := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for malformed json, got %d", w.Code)
	}
}

func TestHistoryAndStatsHandlers(t *testing.T) {
	srv, db := setupTestServer(t, "")
	ctx := t.Context()

	// Seed test data
	fasts := []FastRecord{
		{ID: 1, Start: 100, End: 200, DurationHours: 16.0, TargetHours: 16, State: "COMPLETED"},
		{ID: 2, Start: 300, End: 400, DurationHours: 18.0, TargetHours: 16, State: "COMPLETED"},
	}
	if _, err := db.UpsertFasts(ctx, fasts, "Watch", 500); err != nil {
		t.Fatalf("seed failed: %v", err)
	}

	// Test GET /api/fasting/history
	reqHist := httptest.NewRequest("GET", "/api/fasting/history?limit=10", nil)
	wHist := httptest.NewRecorder()
	srv.Handler().ServeHTTP(wHist, reqHist)

	if wHist.Code != http.StatusOK {
		t.Fatalf("history returned %d: %s", wHist.Code, wHist.Body.String())
	}

	var histResp HistoryResponse
	if err := json.Unmarshal(wHist.Body.Bytes(), &histResp); err != nil {
		t.Fatalf("failed to decode history response: %v", err)
	}
	if histResp.Count != 2 || len(histResp.Records) != 2 {
		t.Errorf("expected 2 records, got count=%d len=%d", histResp.Count, len(histResp.Records))
	}

	// Test GET /api/fasting/stats
	reqStats := httptest.NewRequest("GET", "/api/fasting/stats", nil)
	wStats := httptest.NewRecorder()
	srv.Handler().ServeHTTP(wStats, reqStats)

	if wStats.Code != http.StatusOK {
		t.Fatalf("stats returned %d: %s", wStats.Code, wStats.Body.String())
	}

	var statsResp StatsResponse
	if err := json.Unmarshal(wStats.Body.Bytes(), &statsResp); err != nil {
		t.Fatalf("failed to decode stats response: %v", err)
	}
	if statsResp.TotalFasts != 2 || statsResp.TotalHours != 34.0 {
		t.Errorf("unexpected stats: %+v", statsResp)
	}
}

func TestHealthHandler(t *testing.T) {
	srv, _ := setupTestServer(t, "")

	req := httptest.NewRequest("GET", "/healthz", nil)
	w := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("health check returned %d: %s", w.Code, w.Body.String())
	}

	var healthResp HealthResponse
	if err := json.Unmarshal(w.Body.Bytes(), &healthResp); err != nil {
		t.Fatalf("failed to decode health response: %v", err)
	}
	if healthResp.Status != "ok" {
		t.Errorf("expected status 'ok', got %s", healthResp.Status)
	}
}

func TestDashboardHandler(t *testing.T) {
	srv, db := setupTestServer(t, "")
	ctx := t.Context()

	// Seed one fast
	_, err := db.UpsertFasts(ctx, []FastRecord{
		{ID: 999, Start: 1729000000, End: 1729057600, DurationHours: 16.0, TargetHours: 16, State: "COMPLETED"},
	}, "Instinct 3", 1729060000)
	if err != nil {
		t.Fatalf("seed failed: %v", err)
	}

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK from dashboard, got %d: %s", w.Code, w.Body.String())
	}

	contentType := w.Header().Get("Content-Type")
	if !strings.Contains(contentType, "text/html") {
		t.Errorf("expected text/html content-type, got %s", contentType)
	}

	body := w.Body.String()
	if !strings.Contains(body, "Garmin Fasting Dashboard") {
		t.Errorf("dashboard html missing expected title header")
	}
	if !strings.Contains(body, "16.0h") {
		t.Errorf("dashboard html missing expected duration 16.0h")
	}
}

func TestAPIPost_SingleRecord(t *testing.T) {
	srv, db := setupTestServer(t, "")

	// Post single record (smartphone app shortcut / webhook style)
	singlePayload := []byte(`{
		"start": 1729100000,
		"end": 1729157600,
		"duration_hours": 16.0,
		"target_hours": 16,
		"state": "COMPLETED",
		"device": "iPhone 15 Pro"
	}`)

	req := httptest.NewRequest("POST", "/api", bytes.NewReader(singlePayload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 from POST /api, got %d: %s", w.Code, w.Body.String())
	}

	var resp SyncResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if resp.ReceivedCount != 1 {
		t.Errorf("expected received_count 1, got %d", resp.ReceivedCount)
	}

	count, _ := db.Count(t.Context())
	if count != 1 {
		t.Errorf("expected 1 record in db, got %d", count)
	}
}

func TestAPIPost_RawArray(t *testing.T) {
	srv, db := setupTestServer(t, "")

	// Post raw array of records
	arrayPayload := []byte(`[
		{"id": 201, "start": 1729200000, "end": 1729257600, "duration": 16.0, "target": 16, "state": "COMPLETED"},
		{"id": 202, "start": 1729300000, "end": 1729357600, "duration": 16.0, "target": 16, "state": "COMPLETED"}
	]`)

	req := httptest.NewRequest("POST", "/api", bytes.NewReader(arrayPayload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 from POST /api array, got %d: %s", w.Code, w.Body.String())
	}

	var resp SyncResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if resp.ReceivedCount != 2 {
		t.Errorf("expected received_count 2, got %d", resp.ReceivedCount)
	}

	count, _ := db.Count(t.Context())
	if count != 2 {
		t.Errorf("expected 2 records in db, got %d", count)
	}
}

func TestAPISummaryHandler(t *testing.T) {
	srv, _ := setupTestServer(t, "")

	req := httptest.NewRequest("GET", "/api", nil)
	w := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 from GET /api, got %d: %s", w.Code, w.Body.String())
	}

	var summary APISummaryResponse
	if err := json.Unmarshal(w.Body.Bytes(), &summary); err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if summary.Status != "ok" {
		t.Errorf("expected status 'ok', got %s", summary.Status)
	}
	if summary.Documentation == "" {
		t.Errorf("expected documentation string in api summary")
	}
}

func TestAuth_AlternativeMethods(t *testing.T) {
	const secret = "super-secret"
	srv, _ := setupTestServer(t, secret)

	payload := []byte(`{"start": 100, "end": 200, "duration_hours": 16, "target_hours": 16, "state": "COMPLETED"}`)

	// 1. X-Api-Key header
	req1 := httptest.NewRequest("POST", "/api", bytes.NewReader(payload))
	req1.Header.Set("X-Api-Key", secret)
	w1 := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w1, req1)
	if w1.Code != http.StatusOK {
		t.Errorf("expected 200 with X-Api-Key, got %d", w1.Code)
	}

	// 2. Query param ?token=
	req2 := httptest.NewRequest("POST", "/api?token="+secret, bytes.NewReader(payload))
	w2 := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w2, req2)
	if w2.Code != http.StatusOK {
		t.Errorf("expected 200 with ?token=, got %d", w2.Code)
	}
}
