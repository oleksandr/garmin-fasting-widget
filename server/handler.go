package main

import (
	"bytes"
	"cmp"
	"crypto/subtle"
	"encoding/json/v2"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// Server encapsulates the HTTP handler and application dependencies.
type Server struct {
	store     Store
	authToken string
	startTime time.Time
	mux       *http.ServeMux
}

// NewServer creates a new Server instance with configured routes.
func NewServer(store Store, authToken string) *Server {
	s := &Server{
		store:     store,
		authToken: strings.TrimSpace(authToken),
		startTime: time.Now(),
		mux:       http.NewServeMux(),
	}
	s.routes()
	return s
}

func (s *Server) routes() {
	// Dashboard endpoint
	s.mux.HandleFunc("GET /{$}", s.handleDashboard)

	// Main API endpoints
	s.mux.HandleFunc("POST /api", s.handleSync)
	s.mux.HandleFunc("GET /api", s.handleAPISummary)

	// Garmin & v1 compatibility endpoints
	s.mux.HandleFunc("POST /api/fasting/sync", s.handleSync)
	s.mux.HandleFunc("POST /api/v1/sync", s.handleSync)
	s.mux.HandleFunc("GET /api/fasting/history", s.handleHistory)
	s.mux.HandleFunc("GET /api/v1/history", s.handleHistory)
	s.mux.HandleFunc("GET /api/fasting/stats", s.handleStats)
	s.mux.HandleFunc("GET /api/v1/stats", s.handleStats)

	// Health check endpoints
	s.mux.HandleFunc("GET /healthz", s.handleHealth)
	s.mux.HandleFunc("GET /api/fasting/health", s.handleHealth)
}

func (s *Server) Handler() http.Handler {
	return loggingMiddleware(s.mux)
}

func (s *Server) checkAuth(r *http.Request) bool {
	if s.authToken == "" {
		return true
	}

	token := ""
	if authHeader := r.Header.Get("Authorization"); authHeader != "" {
		if t, ok := strings.CutPrefix(authHeader, "Bearer "); ok {
			token = strings.TrimSpace(t)
		}
	}
	if token == "" {
		token = strings.TrimSpace(r.Header.Get("X-Api-Key"))
	}
	if token == "" {
		token = strings.TrimSpace(r.URL.Query().Get("token"))
	}
	if token == "" {
		token = strings.TrimSpace(r.URL.Query().Get("key"))
	}

	if token == "" {
		return false
	}

	return subtle.ConstantTimeCompare([]byte(token), []byte(s.authToken)) == 1
}

func (s *Server) handleDashboard(w http.ResponseWriter, r *http.Request) {
	data, err := s.store.GetDashboardData(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "database error: "+err.Error())
		return
	}

	data.ServerUptime = time.Since(s.startTime).Truncate(time.Second).String()
	data.ServerTime = time.Now().Format("2006-01-02 15:04:05 MST")
	renderDashboard(w, data)
}

func (s *Server) handleAPISummary(w http.ResponseWriter, r *http.Request) {
	if !s.checkAuth(r) {
		writeError(w, http.StatusUnauthorized, "invalid or missing authorization token")
		return
	}

	stats, err := s.store.GetStats(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "database error: "+err.Error())
		return
	}

	recent, err := s.store.GetHistory(r.Context(), 5)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "database error: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, APISummaryResponse{
		Status:        "ok",
		ServerTime:    time.Now().Format(time.RFC3339),
		Uptime:        time.Since(s.startTime).Truncate(time.Second).String(),
		Stats:         stats,
		RecentFasts:   recent,
		Documentation: "POST JSON fasting records to /api or /api/fasting/sync. View visual dashboard at /",
	})
}

func (s *Server) handleSync(w http.ResponseWriter, r *http.Request) {
	if !s.checkAuth(r) {
		writeError(w, http.StatusUnauthorized, "invalid or missing authorization token")
		return
	}

	// Read body up to 1MB
	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, 1<<20))
	if err != nil {
		writeError(w, http.StatusBadRequest, "failed to read request body: "+err.Error())
		return
	}

	trimmed := bytes.TrimSpace(body)
	if len(trimmed) == 0 {
		writeError(w, http.StatusBadRequest, "request body cannot be empty")
		return
	}

	var fasts []FastRecord
	device := cmp.Or(r.URL.Query().Get("device"), r.Header.Get("X-Device-Name"), "smartphone")
	syncedAt := time.Now().Unix()

	if trimmed[0] == '[' {
		// Array payload: [ { "start": ... }, ... ]
		if err := json.Unmarshal(trimmed, &fasts); err != nil {
			writeError(w, http.StatusBadRequest, "invalid json array: "+err.Error())
			return
		}
	} else if trimmed[0] == '{' {
		// Object payload: FlexibleSyncRequest (supports batch, Garmin sync, or single record)
		var req FlexibleSyncRequest
		if err := json.Unmarshal(trimmed, &req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid json object: "+err.Error())
			return
		}
		fasts = req.ExtractFasts()
		if req.Device != "" {
			device = req.Device
		}
		if req.SyncedAt > 0 {
			syncedAt = req.SyncedAt
		}
	} else {
		writeError(w, http.StatusBadRequest, "invalid json format: expected object or array")
		return
	}

	if len(fasts) == 0 {
		writeError(w, http.StatusBadRequest, "no fasting records found in request")
		return
	}

	count, err := s.store.UpsertFasts(r.Context(), fasts, device, syncedAt)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "database error: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, SyncResponse{
		Status:        "success",
		ReceivedCount: count,
		Message:       "fasting records synced successfully",
	})
}

func (s *Server) handleHistory(w http.ResponseWriter, r *http.Request) {
	if !s.checkAuth(r) {
		writeError(w, http.StatusUnauthorized, "invalid or missing authorization token")
		return
	}

	limit := 30
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsed, err := strconv.Atoi(limitStr); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	records, err := s.store.GetHistory(r.Context(), limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "database error: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, HistoryResponse{
		Status:  "success",
		Count:   len(records),
		Records: records,
	})
}

func (s *Server) handleStats(w http.ResponseWriter, r *http.Request) {
	if !s.checkAuth(r) {
		writeError(w, http.StatusUnauthorized, "invalid or missing authorization token")
		return
	}

	stats, err := s.store.GetStats(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "database error: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, stats)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	count, _ := s.store.Count(r.Context())
	writeJSON(w, http.StatusOK, HealthResponse{
		Status:     "ok",
		Uptime:     time.Since(s.startTime).Truncate(time.Second).String(),
		TotalFasts: count,
	})
}

func writeJSON(w http.ResponseWriter, statusCode int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(statusCode)
	_ = json.MarshalWrite(w, v)
}

func writeError(w http.ResponseWriter, statusCode int, message string) {
	writeJSON(w, statusCode, ErrorResponse{
		Status:  "error",
		Message: message,
	})
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(rw, r)
		log.Printf("%s %s %d %s", r.Method, r.URL.Path, rw.statusCode, time.Since(start))
	})
}
