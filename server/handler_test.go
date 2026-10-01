package main

import (
	"bytes"
	"encoding/json/v2"
	"net/http"
	"net/http/httptest"
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
