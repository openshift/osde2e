package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/openshift/osde2e/pkg/dashboard/collectors"
	"github.com/openshift/osde2e/pkg/dashboard/config"
	"github.com/openshift/osde2e/pkg/dashboard/handlers"
	"github.com/openshift/osde2e/pkg/dashboard/models"
	"github.com/openshift/osde2e/pkg/dashboard/store"
)

// Server represents the dashboard HTTP server
type Server struct {
	config              *config.Config
	reserveCollector    *collectors.ReserveCollector
	usageCollector      *collectors.UsageCollector
	testResultCollector *collectors.TestResultsCollector
	deliverableCollector *collectors.OperatorStatusCollector
	store                *store.Store // optional; when set, deliverables/history served from DB
	mux                 *http.ServeMux
}

// NewServer creates a new dashboard server instance
func NewServer(cfg *config.Config) (*Server, error) {
	// Initialize collectors
	reserveCollector, err := collectors.NewReserveCollector(cfg.OCMEnvironments()...)
	if err != nil {
		log.Printf("Warning: Failed to initialize reserve collector: %v", err)
		reserveCollector = nil
	}

	usageCollector, err := collectors.NewUsageCollector(cfg.OCMEnvironments()...)
	if err != nil {
		log.Printf("Warning: Failed to initialize usage collector: %v", err)
		usageCollector = nil
	}

	var testResultCollector *collectors.TestResultsCollector
	var deliverableCollector *collectors.OperatorStatusCollector
	if cfg.S3Bucket != "" {
		testResultCollector, err = collectors.NewTestResultsCollector(cfg.S3Bucket, cfg.S3Region)
		if err != nil {
			log.Printf("Warning: Failed to initialize test results collector: %v", err)
			testResultCollector = nil
		}

		deliverableCollector, err = collectors.NewOperatorStatusCollector(cfg.S3Bucket, cfg.S3Region, cfg.LookbackDays)
		if err != nil {
			log.Printf("Warning: Failed to initialize deliverable status collector: %v", err)
			deliverableCollector = nil
		}
	}

	srv := &Server{
		config:               cfg,
		reserveCollector:     reserveCollector,
		usageCollector:       usageCollector,
		testResultCollector:  testResultCollector,
		deliverableCollector: deliverableCollector,
		mux:                  http.NewServeMux(),
	}

	// Setup routes
	srv.setupRoutes()

	return srv, nil
}

// setupRoutes configures all HTTP routes
func (s *Server) setupRoutes() {
	// HTML pages
	s.mux.HandleFunc("/", s.handleRedirect)
	s.mux.HandleFunc("/dashboard", s.handleDashboard)
	s.mux.HandleFunc("/dashboard/reserves", s.handleReservesPage)
	s.mux.HandleFunc("/dashboard/usage", s.handleUsagePage)
	s.mux.HandleFunc("/dashboard/deliverables", s.handleDeliverablesPage)
	s.mux.HandleFunc("/dashboard/deliverables/", s.handlePipelineDetailPage)
	s.mux.HandleFunc("/dashboard/analysis", s.handleAnalysisPage)

	// API endpoints
	s.mux.HandleFunc("/api/v1/reserves", s.handleReservesAPI)
	s.mux.HandleFunc("/api/v1/usage", s.handleUsageAPI)
	s.mux.HandleFunc("/api/v1/overview", s.handleOverviewAPI)
	s.mux.HandleFunc("/api/v1/deliverables", s.handleDeliverablesAPI)

	// Health check
	s.mux.HandleFunc("/health", s.handleHealth)
}

// WithStore attaches a SQLite store to the server.
// When set, the deliverables overview and pipeline-detail pages read from the DB
// instead of making live S3 API calls.
func (s *Server) WithStore(st *store.Store) {
	s.store = st
}

// Start starts the HTTP server and blocks until ctx is cancelled, then shuts down gracefully.
func (s *Server) Start(addr string, ctx context.Context) error {
	srv := &http.Server{Addr: addr, Handler: s.mux}

	go func() {
		<-ctx.Done()
		log.Printf("Shutting down dashboard server...")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutdownCtx)
	}()

	log.Printf("Starting server on %s", addr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

// handleRedirect redirects root to /dashboard
func (s *Server) handleRedirect(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" || r.URL.Path == "/dashboard" {
		http.Redirect(w, r, "/dashboard/deliverables", http.StatusMovedPermanently)
		return
	}
	http.NotFound(w, r)
}

// handleDashboard serves the main dashboard HTML page
func (s *Server) handleDashboard(w http.ResponseWriter, r *http.Request) {
	overview, err := s.collectOverview()
	if err != nil {
		s.sendError(w, "Failed to collect dashboard data", http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"ActivePage": "dashboard",
		"Overview":   overview,
	}

	s.renderTemplate(w, "dashboard.html", data)
}

// handleReservesPage serves the reserves HTML page
func (s *Server) handleReservesPage(w http.ResponseWriter, r *http.Request) {
	var reserves []models.ClusterReserve

	if s.reserveCollector != nil {
		collected, err := s.reserveCollector.CollectReserves()
		if err != nil {
			log.Printf("Warning: Failed to collect reserves: %v", err)
			reserves = []models.ClusterReserve{}
		} else {
			reserves = collected
		}
	} else {
		reserves = []models.ClusterReserve{}
	}

	data := map[string]interface{}{
		"ActivePage": "reserves",
		"Reserves":   reserves,
	}

	s.renderTemplate(w, "reserves.html", data)
}

// handleUsagePage serves the Clusters page — all osde2e clusters grouped by env.
func (s *Server) handleUsagePage(w http.ResponseWriter, r *http.Request) {
	// EnvOrder defines the display sequence of environments.
	envOrder := []string{"int", "stage", "prod"}

	type EnvClusters struct {
		Env      string
		Clusters []models.ClusterReserve
	}

	var envClusters []EnvClusters

	if s.reserveCollector != nil {
		byEnv, err := s.reserveCollector.CollectClustersPerEnv()
		if err != nil {
			log.Printf("Warning: Failed to collect clusters per env: %v", err)
		} else {
			for _, env := range envOrder {
				if clusters, ok := byEnv[env]; ok {
					envClusters = append(envClusters, EnvClusters{Env: env, Clusters: clusters})
				}
			}
		}
	}

	data := map[string]interface{}{
		"ActivePage":  "usage",
		"EnvClusters": envClusters,
	}

	s.renderTemplate(w, "usage.html", data)
}

// API Handlers

// handleReservesAPI returns cluster reserves as JSON
func (s *Server) handleReservesAPI(w http.ResponseWriter, r *http.Request) {
	if s.reserveCollector == nil {
		s.sendAPIError(w, "Reserve collector not initialized", http.StatusServiceUnavailable)
		return
	}

	reserves, err := s.reserveCollector.CollectReserves()
	if err != nil {
		s.sendAPIError(w, fmt.Sprintf("Failed to collect reserves: %v", err), http.StatusInternalServerError)
		return
	}

	s.sendAPISuccess(w, reserves)
}

// handleUsageAPI returns cluster usage metrics as JSON
func (s *Server) handleUsageAPI(w http.ResponseWriter, r *http.Request) {
	if s.usageCollector == nil {
		s.sendAPIError(w, "Usage collector not initialized", http.StatusServiceUnavailable)
		return
	}

	env := r.URL.Query().Get("environment")
	if env != "" {
		usage, err := s.usageCollector.CollectUsageByEnvironment(env)
		if err != nil {
			s.sendAPIError(w, fmt.Sprintf("Failed to collect usage: %v", err), http.StatusInternalServerError)
			return
		}
		s.sendAPISuccess(w, usage)
		return
	}

	usage, err := s.usageCollector.CollectUsage()
	if err != nil {
		s.sendAPIError(w, fmt.Sprintf("Failed to collect usage: %v", err), http.StatusInternalServerError)
		return
	}

	s.sendAPISuccess(w, usage)
}

// handleDeliverablesPage serves the deliverables pipeline status HTML page.
// When a Store is configured it reads from SQLite (<1ms); otherwise falls back
// to a live S3 scan (slow, legacy path).
func (s *Server) handleDeliverablesPage(w http.ResponseWriter, r *http.Request) {
	var operators []models.OperatorStatus

	if s.store != nil {
		// Fast path: DB read
		result, err := s.store.GetLatest()
		if err != nil {
			log.Printf("Warning: store.GetLatest: %v", err)
			operators = []models.OperatorStatus{}
		} else {
			operators = result
		}
	} else if s.deliverableCollector != nil {
		// Slow path: live S3 scan
		collected, err := s.deliverableCollector.CollectOperatorStatus()
		if err != nil {
			log.Printf("Warning: Failed to collect deliverable status: %v", err)
			operators = []models.OperatorStatus{}
		} else {
			operators = collected
		}
	} else {
		operators = []models.OperatorStatus{}
	}

	data := map[string]interface{}{
		"ActivePage":   "operators",
		"Operators":    operators,
		"Environments": []string{"stage", "integration"},
		"S3Bucket":     s.config.S3Bucket,
	}

	s.renderTemplate(w, "operators.html", data)
}

// handlePipelineDetailPage serves the per-deliverable pipeline history page.
// URL: /dashboard/deliverables/<name>
// When a Store is configured it reads from SQLite (<1ms); otherwise falls back
// to a live S3 scan (slow, legacy path).
func (s *Server) handlePipelineDetailPage(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimPrefix(r.URL.Path, "/dashboard/deliverables/")
	name = strings.TrimSpace(name)
	if name == "" {
		http.Redirect(w, r, "/dashboard/deliverables", http.StatusSeeOther)
		return
	}

	var history *models.PipelineHistory
	var err error

	if s.store != nil {
		// Fast path: DB read
		history, err = s.store.GetHistory(name)
		if err != nil {
			log.Printf("store.GetHistory %s: %v", name, err)
			s.sendError(w, "Failed to load pipeline history", http.StatusInternalServerError)
			return
		}
	} else if s.deliverableCollector != nil {
		// Slow path: live S3 scan
		history, err = s.deliverableCollector.CollectPipelineHistory(name)
		if err != nil {
			log.Printf("Failed to collect pipeline history for %s: %v", name, err)
			s.sendError(w, "Failed to load pipeline history", http.StatusInternalServerError)
			return
		}
	} else {
		history = &models.PipelineHistory{OperatorName: name}
	}

	data := map[string]interface{}{
		"ActivePage": "operators",
		"History":    history,
	}

	s.renderTemplate(w, "pipeline-detail.html", data)
}

// handleAnalysisPage groups all failed runs by AI root cause and renders the analysis page.
func (s *Server) handleAnalysisPage(w http.ResponseWriter, r *http.Request) {
	var groups []models.FailureGroup

	if s.store != nil {
		var err error
		groups, err = s.store.GetFailureGroups()
		if err != nil {
			log.Printf("Warning: GetFailureGroups: %v", err)
			groups = []models.FailureGroup{}
		}
	}

	data := map[string]interface{}{
		"ActivePage": "analysis",
		"Groups":     groups,
	}

	s.renderTemplate(w, "analysis.html", data)
}

// handleDeliverablesAPI returns deliverable status as JSON
func (s *Server) handleDeliverablesAPI(w http.ResponseWriter, r *http.Request) {
	if s.deliverableCollector == nil {
		s.sendAPIError(w, "Deliverable collector not initialized (S3 bucket not configured)", http.StatusServiceUnavailable)
		return
	}

	operators, err := s.deliverableCollector.CollectOperatorStatus()
	if err != nil {
		s.sendAPIError(w, fmt.Sprintf("Failed to collect operator status: %v", err), http.StatusInternalServerError)
		return
	}

	// Optional ?name= filter
	if nameFilter := r.URL.Query().Get("name"); nameFilter != "" {
		filtered := operators[:0]
		for _, op := range operators {
			if op.Name == nameFilter {
				filtered = append(filtered, op)
			}
		}
		operators = filtered
	}

	s.sendAPISuccess(w, operators)
}

// handleOverviewAPI returns dashboard overview data
func (s *Server) handleOverviewAPI(w http.ResponseWriter, r *http.Request) {
	overview, err := s.collectOverview()
	if err != nil {
		s.sendAPIError(w, fmt.Sprintf("Failed to collect overview: %v", err), http.StatusInternalServerError)
		return
	}

	s.sendAPISuccess(w, overview)
}

// handleHealth returns server health status
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	status := models.HealthStatus{
		Status:       "ok",
		Timestamp:    handlers.Now(),
		OCMConnected: s.reserveCollector != nil,
		S3Connected:  s.testResultCollector != nil,
	}

	if !status.OCMConnected || !status.S3Connected {
		status.Status = "degraded"
	}

	s.sendJSON(w, status)
}

// Helper methods

// collectOverview aggregates data from all collectors
func (s *Server) collectOverview() (*models.DashboardOverview, error) {
	overview := &models.DashboardOverview{
		LastUpdated:         handlers.Now(),
		RecentTests:         []models.TestResult{},
		ClusterUsageSummary: []models.ClusterUsage{},
	}

	// Collect reserves
	if s.reserveCollector != nil {
		reserves, err := s.reserveCollector.CollectReserves()
		if err != nil {
			log.Printf("Warning: Failed to collect reserves: %v", err)
		} else {
			overview.TotalReservedClusters = len(reserves)
			overview.ClustersExpiringSoon = s.reserveCollector.CountExpiringSoon(reserves, s.config.ExpirationWarningThreshold)
		}
	}

	// Collect usage
	if s.usageCollector != nil {
		usage, err := s.usageCollector.CollectUsage()
		if err != nil {
			log.Printf("Warning: Failed to collect usage: %v", err)
		} else {
			overview.ClusterUsageSummary = usage
		}
	}

	// Collect recent tests
	if s.testResultCollector != nil {
		tests, err := s.testResultCollector.CollectRecentTests(20) // Last 20 tests
		if err != nil {
			log.Printf("Warning: Failed to collect test results: %v", err)
		} else {
			overview.RecentTests = tests
			overview.ActiveTests = countActiveTests(tests)
			overview.OverallSuccessRate = calculateSuccessRate(tests)
		}
	}

	return overview, nil
}

// sendJSON sends a JSON response
func (s *Server) sendJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("Error encoding JSON: %v", err)
	}
}

// sendAPISuccess sends a successful API response
func (s *Server) sendAPISuccess(w http.ResponseWriter, data interface{}) {
	s.sendJSON(w, models.APIResponse{
		Success: true,
		Data:    data,
	})
}

// sendAPIError sends an API error response
func (s *Server) sendAPIError(w http.ResponseWriter, message string, statusCode int) {
	w.WriteHeader(statusCode)
	s.sendJSON(w, models.APIResponse{
		Success: false,
		Error:   message,
	})
}

// sendError sends an error response
func (s *Server) sendError(w http.ResponseWriter, message string, statusCode int) {
	http.Error(w, message, statusCode)
}

// Helper functions

func countActiveTests(tests []models.TestResult) int {
	// For now, consider tests from the last hour as "active"
	// This can be refined based on actual test execution patterns
	count := 0
	for _, test := range tests {
		if handlers.Now().Sub(test.Timestamp).Hours() < 1 {
			count++
		}
	}
	return count
}

func calculateSuccessRate(tests []models.TestResult) float64 {
	if len(tests) == 0 {
		return 0
	}

	passed := 0
	for _, test := range tests {
		if test.Status == "passed" {
			passed++
		}
	}

	return float64(passed) / float64(len(tests)) * 100
}
