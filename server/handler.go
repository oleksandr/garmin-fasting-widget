package main

import (
	"crypto/subtle"
	"encoding/json/v2"
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
	// Sync endpoints (both Garmin path and standard v1 API path)
	s.mux.HandleFunc("POST /api/fasting/sync", s.handleSync)
	s.mux.HandleFunc("POST /api/v1/sync", s.handleSync)

	// History query endpoints
	s.mux.HandleFunc("GET /api/fasting/history", s.handleHistory)
	s.mux.HandleFunc("GET /api/v1/history", s.handleHistory)

	// Stats query endpoints
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

	authHeader := r.Header.Get("Authorization")
	token, ok := strings.CutPrefix(authHeader, "Bearer ")
	if !ok {
		return false
	}

	return subtle.ConstantTimeCompare([]byte(strings.TrimSpace(token)), []byte(s.authToken)) == 1
}

func (s *Server) handleSync(w http.ResponseWriter, r *http.Request) {
	if !s.checkAuth(r) {
		writeError(w, http.StatusUnauthorized, "invalid or missing authorization token")
		return
	}

	var req SyncRequest
	if err := json.UnmarshalRead(r.Body, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json payload: "+err.Error())
		return
	}

	count, err := s.store.UpsertFasts(r.Context(), req.Fasts, req.Device, req.SyncedAt)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "database error: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, SyncResponse{
		Status:        "success",
		ReceivedCount: count,
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
