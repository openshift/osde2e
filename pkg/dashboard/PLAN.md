# osde2e Dashboard Implementation Plan

**JIRA**: SDCICD-1823
**Goal**: Web service to gather full context of osde2e operations in each environment

## Overview

A Go-based web dashboard with static snapshots that aggregates:
1. Cluster reserve creations (from OCM API)
2. Cluster usage metrics (from OCM cluster properties)
3. Test status - pass/fail (from S3 bucket `osde2e-logs`)

## Technical Stack

- **Backend**: Go HTTP server (standard library)
- **Frontend**: Go templates (html/template) with minimal JavaScript
- **Data Model**: Static snapshots generated on-demand
- **Data Sources**:
  - OCM API (existing provider integration)
  - S3 bucket: `osde2e-logs` in `us-east-1`
  - Cluster properties: `Availability` (reserved/claimed/used)

## Architecture

### Directory Structure

```
cmd/osde2e/dashboard/
  └── cmd.go              # Cobra command with flags

pkg/dashboard/
  ├── PLAN.md             # This file
  ├── server.go           # HTTP server setup
  ├── config/
  │   └── config.go       # Dashboard configuration
  ├── handlers/
  │   ├── dashboard.go    # HTML page handlers
  │   ├── reserves.go     # Cluster reserve API
  │   ├── usage.go        # Cluster usage API
  │   └── tests.go        # Test results API
  ├── collectors/
  │   ├── reserves.go     # OCM reserve queries
  │   ├── usage.go        # OCM usage queries
  │   └── s3tests.go      # S3 test result fetcher
  ├── models/
  │   └── types.go        # Data models
  ├── templates/
  │   ├── dashboard.html  # Main dashboard page
  │   ├── reserves.html   # Reserves view
  │   ├── usage.html      # Usage view
  │   └── tests.html      # Test results view
  └── docs/
      └── README.md       # Usage documentation
```

### API Endpoints

```
GET  /                      → Redirect to /dashboard
GET  /dashboard             → HTML dashboard home page
GET  /dashboard/reserves    → HTML reserves view
GET  /dashboard/usage       → HTML usage view
GET  /dashboard/tests       → HTML test results view

GET  /api/v1/reserves       → JSON list of reserved clusters
GET  /api/v1/usage          → JSON cluster usage by environment
GET  /api/v1/tests          → JSON test results from S3
GET  /api/v1/tests/:job-id  → JSON detailed test results for job
GET  /health                → Health check endpoint
```

## Data Models

### Cluster Reserve
```go
type ClusterReserve struct {
    ID            string    `json:"id"`
    Name          string    `json:"name"`
    State         string    `json:"state"` // ready, installing, pending
    Availability  string    `json:"availability"` // reserved, claimed, used
    Version       string    `json:"version"`
    Region        string    `json:"region"`
    CloudProvider string    `json:"cloud_provider"`
    CreatedAt     time.Time `json:"created_at"`
    ExpiresAt     time.Time `json:"expires_at"`
    Product       string    `json:"product"` // osd, rosa
}
```

### Cluster Usage
```go
type ClusterUsage struct {
    Environment   string              `json:"environment"` // stage, prod, integration
    TotalClusters int                 `json:"total_clusters"`
    ByState       map[string]int      `json:"by_state"` // ready: 5, installing: 2
    ByAvailability map[string]int     `json:"by_availability"` // reserved: 3, claimed: 2, used: 1
    LastUpdated   time.Time           `json:"last_updated"`
}
```

### Test Result
```go
type TestResult struct {
    JobID         string    `json:"job_id"`
    JobName       string    `json:"job_name"`
    Component     string    `json:"component"`
    Date          string    `json:"date"`
    Status        string    `json:"status"` // passed, failed, error
    TotalTests    int       `json:"total_tests"`
    PassedTests   int       `json:"passed_tests"`
    FailedTests   int       `json:"failed_tests"`
    SkippedTests  int       `json:"skipped_tests"`
    Duration      float64   `json:"duration_seconds"`
    S3Path        string    `json:"s3_path"`
    LogURL        string    `json:"log_url"`
    JUnitXMLURL   string    `json:"junit_xml_url"`
    Timestamp     time.Time `json:"timestamp"`
}
```

## Data Collection

### 1. Cluster Reserves (OCM API)

**Source**: `pkg/common/providers/ocmprovider/cluster.go:QueryReserve()`

Query:
```
cloud_provider.id='<provider>'
AND region.id='<region>'
AND properties.MadeByOSDe2e='true'
AND product.id='<product>'
AND properties.Availability like 'reserved%'
AND version.id like 'openshift-v<version>%'
AND state in ('ready','pending','installing')
```

### 2. Cluster Usage (OCM API)

**Source**: OCM Clusters API with property filtering

Track clusters by:
- Availability property: `reserved`, `claimed`, `used`
- Environment (from provider env setting)
- State: `ready`, `installing`, `pending`, etc.

### 3. Test Results (S3)

**Source**: S3 bucket `osde2e-logs` in `us-east-1`

Path structure: `test-results/<component>/<date>/<job-id>/`

Files to parse:
- `junit*.xml` - JUnit XML test results
- `test_output.log` - Full test logs
- `summary.log` - Test summary

**Existing S3 Integration**: `pkg/common/aws/s3.go`

## CLI Usage

```bash
# Start dashboard server (default port 8080)
osde2e dashboard

# Custom port
osde2e dashboard --port 9000

# Specify environment
osde2e dashboard --environment production

# Custom S3 bucket
osde2e dashboard --s3-bucket osde2e-logs-custom

# Help
osde2e dashboard --help
```

## Dashboard Views

### Main Dashboard (`/dashboard`)
- **Overview Cards**:
  - Total reserved clusters
  - Active tests running
  - Overall test success rate
  - Clusters expiring soon (< 2 hours)
- **Recent Test Results** (last 20):
  - Job name, status, duration, timestamp
  - Pass/fail counts with visual indicators
  - Links to detailed logs
- **Cluster Usage Chart**:
  - Simple HTML/CSS bar chart showing reserved vs claimed vs used

### Reserves View (`/dashboard/reserves`)
- **Filterable Table**:
  - Filter by: state, version, region, cloud provider
  - Sort by: expiration time, created time
  - Columns: ID, Name, State, Availability, Version, Region, Expires At
  - Status indicators (color-coded)
  - Expiration warnings (red if < 2 hours)

### Usage View (`/dashboard/usage`)
- **Environment Breakdown**:
  - Clusters by environment (stage, prod, integration)
  - State distribution (pie chart using HTML/CSS)
  - Availability lifecycle tracking
- **Historical Trends**:
  - Simple time-series showing cluster count over time
  - Peak usage times

### Test Results View (`/dashboard/tests`)
- **Test Job Listings**:
  - Filter by: component, date range, status
  - Sort by: timestamp, duration, failure count
  - Columns: Job ID, Component, Status, Tests (Pass/Fail/Skip), Duration, Timestamp
- **Failure Details**:
  - Expandable rows showing failed test names
  - Links to full logs in S3
  - Quick access to JUnit XML

## Implementation Phases

### Phase 1: Foundation ✓
- [x] Research existing osde2e architecture
- [x] Design data models and API specification
- [ ] Create dashboard command structure
- [ ] Define configuration options

### Phase 2: Data Collection
- [ ] Implement OCM cluster reserve collector
- [ ] Implement OCM cluster usage collector
- [ ] Implement S3 test results collector
- [ ] Add data models and types

### Phase 3: API Layer
- [ ] Create HTTP server with routing
- [ ] Implement API handlers (reserves, usage, tests)
- [ ] Add health check endpoint
- [ ] Handle errors and edge cases

### Phase 4: Frontend
- [ ] Create base HTML template
- [ ] Build dashboard view
- [ ] Build reserves view
- [ ] Build usage view
- [ ] Build test results view
- [ ] Add minimal CSS styling

### Phase 5: Testing & Documentation
- [ ] Add unit tests for collectors
- [ ] Add unit tests for handlers
- [ ] Add integration tests
- [ ] Create usage documentation
- [ ] Add inline code documentation

## Configuration

Dashboard will use existing osde2e config patterns:

```go
// Dashboard configuration keys
const (
    DashboardPort        = "dashboard.port"         // default: 8080
    DashboardS3Bucket    = "dashboard.s3Bucket"     // default: osde2e-logs
    DashboardS3Region    = "dashboard.s3Region"     // default: us-east-1
    DashboardEnvironment = "dashboard.environment"   // default: all
    DashboardRefreshInterval = "dashboard.refreshInterval" // seconds, default: 300
)
```

## Dependencies

All dependencies already exist in osde2e:
- OCM SDK: `github.com/openshift-online/ocm-sdk-go`
- AWS SDK: `github.com/aws/aws-sdk-go`
- Cobra: `github.com/spf13/cobra`
- Viper: Used via `pkg/common/concurrentviper`

## Testing Strategy

1. **Unit Tests**:
   - Collectors: Mock OCM/S3 responses
   - Handlers: Test HTTP responses
   - Models: Validate data transformations

2. **Integration Tests**:
   - End-to-end API tests
   - Template rendering tests
   - S3 bucket access (using test bucket)

3. **Manual Testing**:
   - UI/UX validation
   - Cross-browser compatibility
   - Performance with large datasets

## Security Considerations

- Use existing AWS credentials (via `CcsAwsSession`)
- Use existing OCM authentication
- No additional secrets required
- Read-only access to S3 and OCM
- Rate limiting on API endpoints
- Input validation on query parameters

## Future Enhancements (Out of Scope)

- Real-time updates via WebSocket
- Historical data storage (database)
- Advanced filtering and search
- Prometheus metrics export
- Alerting for expiring clusters
- GraphQL API
- React/Vue.js frontend