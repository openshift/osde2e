# osde2e Dashboard

**JIRA**: SDCICD-1823

A web dashboard for monitoring osde2e operations across environments, providing visibility into cluster reserves, usage metrics, and test results.

## Features

- **Cluster Reserve Tracking**: View all reserved clusters from OCM with status, expiration, and availability
- **Cluster Usage Metrics**: Aggregate cluster usage by environment, state, and cloud provider
- **Test Results**: Browse recent test executions from S3 with pass/fail status and logs
- **REST API**: JSON endpoints for programmatic access to all data
- **Static Snapshots**: On-demand data retrieval (no polling/websockets)

## Architecture

### Components

```
pkg/dashboard/
├── models/          # Data models (ClusterReserve, TestResult, etc.)
├── config/          # Dashboard configuration (reuses common config)
├── collectors/      # Data collectors for OCM and S3
│   ├── reserves.go  # OCM cluster reserve collector
│   ├── usage.go     # OCM cluster usage aggregator
│   └── s3tests.go   # S3 test results parser
├── server/          # HTTP server and routing
│   └── server.go    # Main server with API handlers
├── handlers/        # HTTP handlers and utilities
│   └── utils.go     # Helper functions
├── templates/       # HTML templates (TODO)
└── docs/            # API documentation (TODO)

cmd/osde2e/dashboard/
└── cmd.go           # CLI command with flags
```

### Data Sources

1. **OCM API** (via existing `ocmprovider.OCMProvider`)
   - Cluster reserves with `Availability=reserved`
   - Cluster properties and metadata
   - State tracking (ready, installing, pending)

2. **S3 Bucket** `osde2e-logs` (via existing `aws.CcsAwsSession`)
   - Path: `test-results/<component>/<date>/<job-id>/`
   - JUnit XML test results
   - Test output logs

## Usage

### Start the Dashboard

```bash
# Basic usage (uses defaults)
osde2e dashboard

# Custom port
osde2e dashboard --port 9000

# Filter by environment
osde2e dashboard --environment production

# Limit test results
osde2e dashboard --max-results 50

# With configuration
osde2e dashboard --configs prod --secret-locations /path/to/secrets
```

### Required Environment Variables

```bash
# OCM Configuration
export OCM_CONFIG=/path/to/ocm.json

# AWS Configuration (for S3 access)
export AWS_ACCESS_KEY_ID=your_key
export AWS_SECRET_ACCESS_KEY=your_secret
export LOG_BUCKET=osde2e-logs  # Optional, defaults to osde2e-logs
```

### API Endpoints

All endpoints return JSON responses.

#### Dashboard Overview
```
GET /api/v1/overview
```
Returns aggregated dashboard data including reserves, usage, and recent tests.

#### Cluster Reserves
```
GET /api/v1/reserves
```
Lists all reserved clusters from OCM.

Response:
```json
{
  "success": true,
  "data": [
    {
      "id": "cluster-123",
      "name": "osde2e-abc",
      "state": "ready",
      "availability": "reserved",
      "version": "openshift-v4.14.0",
      "region": "us-east-1",
      "cloud_provider": "aws",
      "created_at": "2026-04-30T10:00:00Z",
      "expires_at": "2026-05-01T10:00:00Z",
      "product": "rosa"
    }
  ]
}
```

#### Cluster Usage
```
GET /api/v1/usage
GET /api/v1/usage?environment=production
```
Returns cluster usage metrics aggregated by environment.

Response:
```json
{
  "success": true,
  "data": [
    {
      "environment": "production",
      "total_clusters": 25,
      "by_state": {
        "ready": 20,
        "installing": 3,
        "pending": 2
      },
      "by_availability": {
        "reserved": 10,
        "claimed": 8,
        "used": 7
      },
      "last_updated": "2026-04-30T12:00:00Z"
    }
  ]
}
```

#### Test Results
```
GET /api/v1/tests
GET /api/v1/tests/{job-id}
```
Lists recent test results or retrieves a specific test by job ID.

Response:
```json
{
  "success": true,
  "data": [
    {
      "job_id": "abc123",
      "job_name": "periodic-ci-openshift-osde2e",
      "component": "osd-example-operator",
      "date": "2026-04-30",
      "status": "passed",
      "total_tests": 50,
      "passed_tests": 48,
      "failed_tests": 2,
      "skipped_tests": 0,
      "duration_seconds": 1234.5,
      "s3_path": "test-results/osd-example-operator/2026-04-30/abc123",
      "log_url": "https://s3.amazonaws.com/...",
      "junit_xml_url": "https://s3.amazonaws.com/...",
      "timestamp": "2026-04-30T11:30:00Z"
    }
  ]
}
```

#### Health Check
```
GET /health
```
Returns server health status.

Response:
```json
{
  "status": "ok",
  "timestamp": "2026-04-30T12:00:00Z",
  "ocm_connected": true,
  "s3_connected": true
}
```

## Configuration

The dashboard reuses existing osde2e configuration:

| Config Key | Environment Variable | Default | Description |
|------------|---------------------|---------|-------------|
| `dashboard.port` | - | `8080` | HTTP server port |
| `dashboard.environment` | - | `all` | Filter environment |
| `dashboard.maxTestResults` | - | `100` | Max test results to return |
| `tests.logBucket` | `LOG_BUCKET` | `osde2e-logs` | S3 bucket for test results |
| `config.aws.region` | `AWS_REGION` | `us-east-1` | S3 bucket region |
| `ocmConfig` | `OCM_CONFIG` | - | Path to OCM config file |

## Implementation Status

### Completed ✅
- [x] Data models and types
- [x] Configuration management (reuses common config)
- [x] OCM cluster reserve collector
- [x] OCM cluster usage collector
- [x] S3 test results collector (with JUnit XML parsing)
- [x] HTTP server with routing
- [x] REST API handlers
- [x] Dashboard command (CLI)
- [x] Integration with main osde2e command

### TODO 🚧
- [ ] HTML templates for web UI
- [ ] CSS styling for dashboard pages
- [ ] Unit tests for collectors
- [ ] Unit tests for handlers
- [ ] Integration tests
- [ ] Build verification
- [ ] Deployment documentation

## Development

### Project Structure

The dashboard follows osde2e patterns:
- Reuses existing AWS and OCM connections
- Uses viper for configuration
- Follows cobra command structure
- Integrates with existing providers

### Adding New Features

1. **New Data Source**: Add collector in `collectors/`
2. **New API Endpoint**: Add handler in `server/server.go`
3. **New Configuration**: Add to `config/config.go`
4. **New Model**: Add to `models/types.go`

### Testing

```bash
# Run unit tests (when implemented)
go test ./pkg/dashboard/...

# Run with test configuration
osde2e dashboard --configs test --port 8080

# Test API endpoints
curl http://localhost:8080/api/v1/reserves
curl http://localhost:8080/api/v1/usage
curl http://localhost:8080/api/v1/tests
curl http://localhost:8080/health
```

## Next Steps

1. **Build Verification**: Test compilation and fix any errors
2. **HTML Templates**: Create Go templates for web UI
3. **Testing**: Add comprehensive unit and integration tests
4. **Documentation**: Complete API documentation
5. **Deployment**: Add deployment instructions and examples

## Contributing

When adding new features:
1. Follow existing code patterns
2. Reuse common osde2e utilities
3. Add appropriate error handling
4. Update this README
5. Add tests for new functionality

## Troubleshooting

### OCM Connection Issues
```
Warning: OCM_CONFIG not set. OCM features may not work.
```
Solution: Set `OCM_CONFIG` environment variable to your ocm.json path.

### S3 Access Issues
```
Warning: LOG_BUCKET not set. S3 test results will not be available.
```
Solution: Set `LOG_BUCKET` and AWS credentials.

### No Data Returned
Check that:
- OCM config is valid and accessible
- AWS credentials have S3 read access
- Clusters exist with `MadeByOSDe2e=true` property
- Test results exist in S3 bucket

## License

Same as osde2e project.