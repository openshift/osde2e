# Template Fix - Empty State Handling

**Issue**: Dashboard showing "No cluster usage data available" even when data exists

**Root Cause**: Template conditionals checking for nil slices instead of empty slices

## Problem

Original template code:
```go
{{if .Overview.ClusterUsageSummary}}
  <!-- Show table -->
{{else}}
  <!-- Show "No data" message -->
{{end}}
```

This checks if the slice is non-nil, but an **empty slice** (length 0) is not nil, so:
- Empty slice `[]` → Truthy → Shows empty table (confusing)
- Nil slice → Falsy → Shows "No data" message (correct)

## Solution

Updated all templates to check length:
```go
{{if gt (len .Overview.ClusterUsageSummary) 0}}
  <!-- Show table -->
{{else}}
  <!-- Show "No data" message -->
{{end}}
```

This properly checks if the slice has elements:
- Empty slice `[]` → Length 0 → Shows "No data" message (correct)
- Non-empty slice → Length > 0 → Shows table (correct)
- Nil slice → Length 0 → Shows "No data" message (correct)

## Files Updated

1. **dashboard.html**
   - `{{if gt (len .Overview.RecentTests) 0}}` - Recent test results
   - `{{if gt (len .Overview.ClusterUsageSummary) 0}}` - Cluster usage summary

2. **reserves.html**
   - `{{if gt (len .Reserves) 0}}` - Reserved clusters table

3. **usage.html**
   - `{{if gt (len .Usage) 0}}` - Usage metrics by environment

4. **tests.html**
   - `{{if gt (len .Tests) 0}}` - Test results table

## Testing

### Before Fix
- No osde2e clusters → Shows empty table header with no rows (confusing)
- Has osde2e clusters → Shows table with data (works)

### After Fix
- No osde2e clusters → Shows "No data available" message (correct)
- Has osde2e clusters → Shows table with data (correct)

## Additional Improvements

Added helpful context to empty state messages:

**Dashboard page**:
```html
<p>No cluster usage data available</p>
<p class="text-muted">Clusters made by osde2e (with MadeByOSDe2e=true) will appear here</p>
```

**Reserves page**:
```html
<p>No reserved clusters found</p>
<p class="text-muted">Clusters with Availability=reserved will appear here</p>
```

**Tests page**:
```html
<p>No test results found</p>
<p class="text-muted">Test results from S3 bucket will appear here</p>
```

## Best Practice

When working with Go templates and slices, always use:
```go
{{if gt (len .SliceName) 0}}
```

Instead of:
```go
{{if .SliceName}}
```

This ensures proper handling of:
- Nil slices
- Empty slices
- Non-empty slices

---

**Status**: ✅ Fixed
**Date**: April 30, 2026
