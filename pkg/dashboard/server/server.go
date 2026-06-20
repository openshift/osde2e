package server

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	awss3 "github.com/aws/aws-sdk-go/service/s3"
	junit "github.com/joshdk/go-junit"
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
	deliverableCollector *collectors.DeliverableCollector
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
	var deliverableCollector *collectors.DeliverableCollector
	if cfg.S3Bucket != "" {
		testResultCollector, err = collectors.NewTestResultsCollector(cfg.S3Bucket, cfg.S3Region)
		if err != nil {
			log.Printf("Warning: Failed to initialize test results collector: %v", err)
			testResultCollector = nil
		}

		deliverableCollector, err = collectors.NewDeliverableCollector(cfg.S3Bucket, cfg.S3Region, cfg.LookbackDays)
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
	s.mux.HandleFunc("/dashboard/usage", s.handleUsagePage)
	s.mux.HandleFunc("/dashboard/pipelines", s.handleDeliverablesPage)
	s.mux.HandleFunc("/dashboard/pipelines/", s.handlePipelineDetailPage)
	s.mux.HandleFunc("/dashboard/analysis", s.handleAnalysisPage)

	// API endpoints
	s.mux.HandleFunc("/api/v1/reserves", s.handleReservesAPI)
	s.mux.HandleFunc("/api/v1/usage", s.handleUsageAPI)
	s.mux.HandleFunc("/api/v1/overview", s.handleOverviewAPI)
	s.mux.HandleFunc("/api/v1/deliverables", s.handleDeliverablesAPI)

	// S3 object proxy (streams objects server-side, no presigned URL expiry)
	s.mux.HandleFunc("/dashboard/s3", s.handleS3Proxy)

	// JUnit XML viewer
	s.mux.HandleFunc("/dashboard/junit", s.handleJUnitReport)

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
	srv := &http.Server{
		Addr:              addr,
		Handler:           s.mux,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

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
		http.Redirect(w, r, "/dashboard/pipelines", http.StatusMovedPermanently)
		return
	}
	http.NotFound(w, r)
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
	var deliverables []models.DeliverableStatus

	if s.store != nil {
		// Fast path: DB read
		result, err := s.store.GetLatest()
		if err != nil {
			log.Printf("Warning: store.GetLatest: %v", err)
			deliverables = []models.DeliverableStatus{}
		} else {
			deliverables = result
		}
	} else if s.deliverableCollector != nil {
		// Slow path: live S3 scan
		collected, err := s.deliverableCollector.CollectDeliverables()
		if err != nil {
			log.Printf("Warning: Failed to collect deliverable status: %v", err)
			deliverables = []models.DeliverableStatus{}
		} else {
			deliverables = collected
		}
	} else {
		deliverables = []models.DeliverableStatus{}
	}

	data := map[string]interface{}{
		"ActivePage":   "deliverables",
		"Deliverables":  deliverables,
		"Environments": []string{"stage", "integration"},
		"S3Bucket":     s.config.S3Bucket,
	}

	s.renderTemplate(w, "deliverables.html", data)
}

// handlePipelineDetailPage serves the per-deliverable pipeline history page.
// URL: /dashboard/pipelines/<name>
// When a Store is configured it reads from SQLite (<1ms); otherwise falls back
// to a live S3 scan (slow, legacy path).
func (s *Server) handlePipelineDetailPage(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimPrefix(r.URL.Path, "/dashboard/pipelines/")
	name = strings.TrimSpace(name)
	if name == "" {
		http.Redirect(w, r, "/dashboard/pipelines", http.StatusSeeOther)
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
		history = &models.PipelineHistory{Name: name}
	}

	data := map[string]interface{}{
		"ActivePage": "deliverables",
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

	deliverables, err := s.deliverableCollector.CollectDeliverables()
	if err != nil {
		s.sendAPIError(w, fmt.Sprintf("Failed to collect deliverable status: %v", err), http.StatusInternalServerError)
		return
	}

	// Optional ?name= filter
	if nameFilter := r.URL.Query().Get("name"); nameFilter != "" {
		filtered := deliverables[:0]
		for _, op := range deliverables {
			if op.Name == nameFilter {
				filtered = append(filtered, op)
			}
		}
		deliverables = filtered
	}

	s.sendAPISuccess(w, deliverables)
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

// handleS3Proxy streams an S3 object through the server using its AWS credentials.
// URL: /dashboard/s3?key=<s3-object-key>
// This avoids presigned URL expiry — the server holds long-lived credentials.
func (s *Server) handleS3Proxy(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")
	if key == "" {
		http.Error(w, "missing key parameter", http.StatusBadRequest)
		return
	}
	if s.deliverableCollector == nil {
		http.Error(w, "S3 not configured", http.StatusServiceUnavailable)
		return
	}

	s3Client, bucket := s.deliverableCollector.S3Client()
	out, err := s3Client.GetObjectWithContext(r.Context(), &awss3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		log.Printf("handleS3Proxy: GetObject %s: %v", key, err)
		http.Error(w, "Failed to fetch object from S3", http.StatusBadGateway)
		return
	}
	defer out.Body.Close()

	if ct := aws.StringValue(out.ContentType); ct != "" {
		w.Header().Set("Content-Type", ct)
	} else if strings.HasSuffix(key, ".log") {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	} else if strings.HasSuffix(key, ".xml") {
		w.Header().Set("Content-Type", "application/xml")
	}
	_, _ = io.Copy(w, out.Body)
}

// handleJUnitReport fetches a JUnit XML from S3 and renders it as HTML.
// URL: /dashboard/junit?key=<s3-object-key>
func (s *Server) handleJUnitReport(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")
	if key == "" {
		http.Error(w, "missing key parameter", http.StatusBadRequest)
		return
	}
	if s.deliverableCollector == nil {
		http.Error(w, "S3 not configured", http.StatusServiceUnavailable)
		return
	}

	s3Client, bucket := s.deliverableCollector.S3Client()
	out, err := s3Client.GetObjectWithContext(r.Context(), &awss3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		log.Printf("handleJUnitReport: GetObject %s: %v", key, err)
		s.sendError(w, "Failed to fetch JUnit XML from S3", http.StatusBadGateway)
		return
	}
	defer out.Body.Close()

	suites, err := junit.IngestReader(out.Body)
	if err != nil {
		log.Printf("handleJUnitReport: parse error: %v", err)
		s.sendError(w, "Failed to parse JUnit XML", http.StatusUnprocessableEntity)
		return
	}

	s.renderTemplate(w, "junit-report.html", map[string]interface{}{
		"ActivePage": "deliverables",
		"Suites":     suites,
	})
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
