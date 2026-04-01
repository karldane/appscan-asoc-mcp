package tools

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/karldane/appscan-asoc-mcp/internal/client"
)

// skipUnlessLive skips the test unless APPSCAN_LIVE_TEST=1 is set.
func skipUnlessLive(t *testing.T) {
	t.Helper()
	if os.Getenv("APPSCAN_LIVE_TEST") != "1" {
		t.Skip("skipping live test (set APPSCAN_LIVE_TEST=1 to run)")
	}
}

// liveClient creates a client from environment variables for live testing.
// Returns nil if credentials are not set.
func liveClient(t *testing.T) *client.Client {
	t.Helper()

	baseURL := os.Getenv("APPSCAN_BASE_URL")
	keyID := os.Getenv("APPSCAN_KEY_ID")
	keySecret := os.Getenv("APPSCAN_KEY_SECRET")

	// Support combined key format
	if keyID == "" && keySecret == "" {
		combined := os.Getenv("APPSCAN_API_KEY")
		if combined != "" {
			// Parse keyid:keysecret
			for i, c := range combined {
				if c == ':' {
					keyID = combined[:i]
					keySecret = combined[i+1:]
					break
				}
			}
		}
	}

	if baseURL == "" || keyID == "" || keySecret == "" {
		t.Skip("APPSCAN_BASE_URL, APPSCAN_KEY_ID, and APPSCAN_KEY_SECRET must be set for live tests")
	}

	// Normalize base URL to include /api/v4 suffix
	if !strings.HasSuffix(baseURL, "/api/v4") {
		if strings.HasSuffix(baseURL, "/") {
			baseURL = baseURL + "api/v4"
		} else {
			baseURL = baseURL + "/api/v4"
		}
	}

	return client.New(baseURL, keyID, keySecret, 30)
}

// findTestAppID returns an application ID from the live API for testing.
// This avoids hardcoding an ID that may not exist in all tenants.
func findTestAppID(t *testing.T, c *client.Client) string {
	t.Helper()

	tool := NewAppsListTool(c, nil)
	result, err := tool.Handle(context.Background(), map[string]interface{}{
		"page_size": float64(5),
	})
	if err != nil {
		t.Fatalf("Failed to list apps for test setup: %v", err)
	}

	var output map[string]interface{}
	if err := json.Unmarshal([]byte(result), &output); err != nil {
		t.Fatalf("Failed to parse apps list: %v", err)
	}

	apps, ok := output["applications"].([]interface{})
	if !ok || len(apps) == 0 {
		t.Skip("No applications found for testing")
	}

	firstApp := apps[0].(map[string]interface{})
	appID, ok := firstApp["id"].(string)
	if !ok || appID == "" {
		t.Skip("Could not extract application ID from response")
	}

	return appID
}

// findTestScanID returns a scan ID from the live API for testing.
func findTestScanID(t *testing.T, c *client.Client, appID string) string {
	t.Helper()

	tool := NewScansListTool(c, nil)
	result, err := tool.Handle(context.Background(), map[string]interface{}{
		"application_id": appID,
		"page_size":      float64(5),
	})
	if err != nil {
		t.Skipf("Could not list scans for app %s: %v", appID, err)
	}

	var output map[string]interface{}
	if err := json.Unmarshal([]byte(result), &output); err != nil {
		t.Skipf("Failed to parse scans list: %v", err)
	}

	scans, ok := output["scans"].([]interface{})
	if !ok || len(scans) == 0 {
		t.Skip("No scans found for testing")
	}

	firstScan := scans[0].(map[string]interface{})
	scanID, ok := firstScan["id"].(string)
	if !ok || scanID == "" {
		t.Skip("Could not extract scan ID from response")
	}

	return scanID
}

// findTestFindingID returns a finding ID from the live API for testing.
func findTestFindingID(t *testing.T, c *client.Client, appID string) string {
	t.Helper()

	tool := NewFindingsListTool(c, nil)
	result, err := tool.Handle(context.Background(), map[string]interface{}{
		"application_id": appID,
		"page_size":      float64(5),
	})
	if err != nil {
		t.Skipf("Could not list findings for app %s: %v", appID, err)
	}

	var output map[string]interface{}
	if err := json.Unmarshal([]byte(result), &output); err != nil {
		t.Skipf("Failed to parse findings list: %v", err)
	}

	findings, ok := output["findings"].([]interface{})
	if !ok || len(findings) == 0 {
		t.Skip("No findings found for testing")
	}

	firstFinding := findings[0].(map[string]interface{})
	findingID, ok := firstFinding["id"].(string)
	if !ok || findingID == "" {
		t.Skip("Could not extract finding ID from response")
	}

	return findingID
}

// ---------------------------------------------------------------------------
// Live endpoint verification tests
// These tests verify that each read-only tool correctly calls the ASoC API.
// They are gated by APPSCAN_LIVE_TEST=1 and use real credentials.
// ---------------------------------------------------------------------------

func TestLive_AppsList(t *testing.T) {
	skipUnlessLive(t)
	c := liveClient(t)

	tool := NewAppsListTool(c, nil)
	result, err := tool.Handle(context.Background(), map[string]interface{}{
		"page_size": float64(5),
	})
	if err != nil {
		t.Fatalf("apps_list failed: %v", err)
	}

	var output map[string]interface{}
	if err := json.Unmarshal([]byte(result), &output); err != nil {
		t.Fatalf("apps_list returned invalid JSON: %v", err)
	}

	apps, ok := output["applications"].([]interface{})
	if !ok {
		t.Fatal("apps_list response missing 'applications' array")
	}
	t.Logf("apps_list: found %d applications", len(apps))
}

func TestLive_AppGet(t *testing.T) {
	skipUnlessLive(t)
	c := liveClient(t)

	appID := findTestAppID(t, c)

	tool := NewAppGetTool(c, nil)
	result, err := tool.Handle(context.Background(), map[string]interface{}{
		"id": appID,
	})
	if err != nil {
		t.Fatalf("app_get failed: %v", err)
	}

	var output map[string]interface{}
	if err := json.Unmarshal([]byte(result), &output); err != nil {
		t.Fatalf("app_get returned invalid JSON: %v", err)
	}

	if output["id"] != appID {
		t.Errorf("app_get returned wrong ID: got %v, want %s", output["id"], appID)
	}
	t.Logf("app_get: %s (%s)", output["name"], appID)
}

func TestLive_ScansList(t *testing.T) {
	skipUnlessLive(t)
	c := liveClient(t)

	appID := findTestAppID(t, c)

	tool := NewScansListTool(c, nil)
	result, err := tool.Handle(context.Background(), map[string]interface{}{
		"application_id": appID,
		"page_size":      float64(5),
	})
	if err != nil {
		t.Fatalf("scans_list failed: %v", err)
	}

	var output map[string]interface{}
	if err := json.Unmarshal([]byte(result), &output); err != nil {
		t.Fatalf("scans_list returned invalid JSON: %v", err)
	}

	scans, ok := output["scans"].([]interface{})
	if !ok {
		t.Fatal("scans_list response missing 'scans' array")
	}
	t.Logf("scans_list: found %d scans for app %s", len(scans), appID)
}

func TestLive_ScanGet(t *testing.T) {
	skipUnlessLive(t)
	c := liveClient(t)

	appID := findTestAppID(t, c)
	scanID := findTestScanID(t, c, appID)

	tool := NewScanGetTool(c, nil)
	result, err := tool.Handle(context.Background(), map[string]interface{}{
		"id": scanID,
	})
	if err != nil {
		t.Fatalf("scan_get failed: %v", err)
	}

	var output map[string]interface{}
	if err := json.Unmarshal([]byte(result), &output); err != nil {
		t.Fatalf("scan_get returned invalid JSON: %v", err)
	}

	if output["id"] != scanID {
		t.Errorf("scan_get returned wrong ID: got %v, want %s", output["id"], scanID)
	}
	t.Logf("scan_get: %s (status: %s)", scanID, output["status"])
}

func TestLive_ScanStatus(t *testing.T) {
	skipUnlessLive(t)
	c := liveClient(t)

	appID := findTestAppID(t, c)
	scanID := findTestScanID(t, c, appID)

	tool := NewScanStatusTool(c, nil)
	result, err := tool.Handle(context.Background(), map[string]interface{}{
		"id": scanID,
	})
	if err != nil {
		t.Fatalf("scan_status failed: %v", err)
	}

	var output map[string]interface{}
	if err := json.Unmarshal([]byte(result), &output); err != nil {
		t.Fatalf("scan_status returned invalid JSON: %v", err)
	}

	t.Logf("scan_status: %s (status: %s)", scanID, output["status"])
}

func TestLive_FindingsList(t *testing.T) {
	skipUnlessLive(t)
	c := liveClient(t)

	appID := findTestAppID(t, c)

	tool := NewFindingsListTool(c, nil)
	result, err := tool.Handle(context.Background(), map[string]interface{}{
		"application_id": appID,
		"page_size":      float64(5),
	})
	if err != nil {
		t.Fatalf("findings_list failed: %v", err)
	}

	var output map[string]interface{}
	if err := json.Unmarshal([]byte(result), &output); err != nil {
		t.Fatalf("findings_list returned invalid JSON: %v", err)
	}

	findings, ok := output["findings"].([]interface{})
	if !ok {
		t.Fatal("findings_list response missing 'findings' array")
	}
	t.Logf("findings_list: found %d findings for app %s", len(findings), appID)
}

func TestLive_FindingsSearch(t *testing.T) {
	skipUnlessLive(t)
	c := liveClient(t)

	appID := findTestAppID(t, c)

	tool := NewFindingsSearchTool(c, nil)
	result, err := tool.Handle(context.Background(), map[string]interface{}{
		"application_id": appID,
		"severity":       "High",
		"page_size":      float64(5),
	})
	if err != nil {
		t.Fatalf("findings_search failed: %v", err)
	}

	var output map[string]interface{}
	if err := json.Unmarshal([]byte(result), &output); err != nil {
		t.Fatalf("findings_search returned invalid JSON: %v", err)
	}

	findings, ok := output["findings"].([]interface{})
	if !ok {
		t.Fatal("findings_search response missing 'findings' array")
	}
	t.Logf("findings_search (High): found %d findings for app %s", len(findings), appID)
}

func TestLive_FindingGet(t *testing.T) {
	skipUnlessLive(t)
	c := liveClient(t)

	appID := findTestAppID(t, c)
	findingID := findTestFindingID(t, c, appID)

	tool := NewFindingGetTool(c, nil)
	result, err := tool.Handle(context.Background(), map[string]interface{}{
		"id": findingID,
	})
	if err != nil {
		t.Fatalf("finding_get failed: %v", err)
	}

	var output map[string]interface{}
	if err := json.Unmarshal([]byte(result), &output); err != nil {
		t.Fatalf("finding_get returned invalid JSON: %v", err)
	}

	if output["id"] != findingID {
		t.Errorf("finding_get returned wrong ID: got %v, want %s", output["id"], findingID)
	}
	t.Logf("finding_get: %s (severity: %s, status: %s)", findingID, output["severity"], output["status"])
}

func TestLive_ReportsList(t *testing.T) {
	skipUnlessLive(t)
	c := liveClient(t)

	appID := findTestAppID(t, c)

	tool := NewReportsListTool(c, nil)
	result, err := tool.Handle(context.Background(), map[string]interface{}{
		"application_id": appID,
		"page_size":      float64(5),
	})
	if err != nil {
		t.Fatalf("reports_list failed: %v", err)
	}

	var output map[string]interface{}
	if err := json.Unmarshal([]byte(result), &output); err != nil {
		t.Fatalf("reports_list returned invalid JSON: %v", err)
	}

	t.Logf("reports_list: returned %d bytes of JSON", len(result))
}

func TestLive_AssetGroupsList(t *testing.T) {
	skipUnlessLive(t)
	c := liveClient(t)

	tool := NewAssetGroupsListTool(c, nil)
	result, err := tool.Handle(context.Background(), map[string]interface{}{
		"page_size": float64(10),
	})
	if err != nil {
		t.Fatalf("asset_groups_list failed: %v", err)
	}

	var output map[string]interface{}
	if err := json.Unmarshal([]byte(result), &output); err != nil {
		t.Fatalf("asset_groups_list returned invalid JSON: %v", err)
	}

	groups, _ := output["asset_groups"].([]interface{})
	t.Logf("asset_groups_list: found %d asset groups", len(groups))
}

func TestLive_PoliciesList(t *testing.T) {
	skipUnlessLive(t)
	c := liveClient(t)

	tool := NewPoliciesListTool(c, nil)
	result, err := tool.Handle(context.Background(), map[string]interface{}{
		"page_size": float64(10),
	})
	if err != nil {
		t.Fatalf("policies_list failed: %v", err)
	}

	var output map[string]interface{}
	if err := json.Unmarshal([]byte(result), &output); err != nil {
		t.Fatalf("policies_list returned invalid JSON: %v", err)
	}

	policies, _ := output["policies"].([]interface{})
	t.Logf("policies_list: found %d policies", len(policies))
}

func TestLive_ComplianceSummary(t *testing.T) {
	skipUnlessLive(t)
	c := liveClient(t)

	appID := findTestAppID(t, c)

	tool := NewComplianceSummaryTool(c, nil)
	result, err := tool.Handle(context.Background(), map[string]interface{}{
		"application_id": appID,
	})

	// Note: compliance_summary endpoint may not be available in all ASoC tenants
	// or may have a different endpoint path. Skip if 404.
	if err != nil {
		if strings.Contains(err.Error(), "HTTP 404") {
			t.Skipf("compliance_summary endpoint not available (HTTP 404): %v", err)
		}
		t.Fatalf("compliance_summary failed: %v", err)
	}

	var output map[string]interface{}
	if err := json.Unmarshal([]byte(result), &output); err != nil {
		t.Fatalf("compliance_summary returned invalid JSON: %v", err)
	}

	t.Logf("compliance_summary: returned %d bytes of JSON for app %s", len(result), appID)
}

func TestLive_FindingsSearch_WithStatus(t *testing.T) {
	skipUnlessLive(t)
	c := liveClient(t)

	appID := findTestAppID(t, c)

	tool := NewFindingsSearchTool(c, nil)
	result, err := tool.Handle(context.Background(), map[string]interface{}{
		"application_id": appID,
		"status":         "Open",
		"page_size":      float64(5),
	})
	if err != nil {
		t.Fatalf("findings_search (status=Open) failed: %v", err)
	}

	var output map[string]interface{}
	if err := json.Unmarshal([]byte(result), &output); err != nil {
		t.Fatalf("findings_search returned invalid JSON: %v", err)
	}

	findings, ok := output["findings"].([]interface{})
	if !ok {
		t.Fatal("findings_search response missing 'findings' array")
	}
	t.Logf("findings_search (Open): found %d findings for app %s", len(findings), appID)
}
