// Health-check URL parsing and validation for krkn-ai config.
package krknai

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// redactURL returns a URL string safe for logging: userinfo and query are stripped.
// Invalid URLs return "<redacted>".
func redactURL(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return "<redacted>"
	}
	u.User = nil
	u.RawQuery = ""
	u.Fragment = ""
	return u.String()
}

// validateHealthCheckURLsReachable performs HTTP GET on each health check URL and returns an error
// if any return non-2xx or are unreachable. URLs in errors are redacted.
func validateHealthCheckURLsReachable(ctx context.Context, apps []map[string]interface{}) error {
	const perRequestTimeout = 10 * time.Second
	client := &http.Client{Timeout: perRequestTimeout}
	var errs []string
	for _, app := range apps {
		name, _ := app["name"].(string)
		rawURL, _ := app["url"].(string)
		if rawURL == "" {
			continue
		}
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
		if err != nil {
			errs = append(errs, fmt.Sprintf("%s (%s): %v", name, redactURL(rawURL), err))
			continue
		}
		resp, err := client.Do(req)
		if err != nil {
			errs = append(errs, fmt.Sprintf("%s (%s): %v", name, redactURL(rawURL), err))
			continue
		}
		_ = resp.Body.Close()
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			errs = append(errs, fmt.Sprintf("%s (%s): HTTP %d", name, redactURL(rawURL), resp.StatusCode))
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("health check URL validation failed: %s", strings.Join(errs, "; "))
	}
	return nil
}

// parseHealthCheckEndpoints parses a comma-separated string of name=url pairs
// into health check application entries for the krkn-ai config. Returns an error
// on the first invalid entry (invalid URL, empty name, unsupported scheme, etc.).
func parseHealthCheckEndpoints(input string) ([]map[string]interface{}, error) {
	var apps []map[string]interface{}
	for _, entry := range strings.Split(input, ",") {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}
		parts := strings.SplitN(entry, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid health-check entry (expected name=url): %q", entry)
		}
		name := strings.TrimSpace(parts[0])
		rawURL := strings.TrimSpace(parts[1])
		if name == "" || rawURL == "" {
			return nil, fmt.Errorf("invalid health-check entry (name and url required): %q", entry)
		}
		u, err := url.Parse(rawURL)
		if err != nil || u.Scheme == "" || u.Host == "" {
			return nil, fmt.Errorf("invalid URL for %q (must include scheme and host, e.g. https://host/path): %q", name, redactURL(rawURL))
		}
		if u.Scheme != "http" && u.Scheme != "https" {
			return nil, fmt.Errorf("unsupported scheme %q for %q (must be http or https)", u.Scheme, name)
		}
		apps = append(apps, map[string]interface{}{
			"name":        name,
			"url":         rawURL,
			"status_code": 200,
			"timeout":     4,
			"interval":    2,
		})
	}
	return apps, nil
}
