package normalize

import (
	"testing"
)

func TestNormalizeSeverity(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Critical", "high"},
		{"High", "high"},
		{"Medium", "medium"},
		{"Low", "low"},
		{"Info", "info"},
		{"Informational", "info"},
		{"Unknown", "unknown"},
	}

	for _, tt := range tests {
		result := normalizeSeverity(tt.input)
		if result != tt.expected {
			t.Errorf("normalizeSeverity(%s) = %s; want %s", tt.input, result, tt.expected)
		}
	}
}

func TestNormalizeStatus(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Ready", "completed"},
		{"Completed", "completed"},
		{"Running", "running"},
		{"Failed", "failed"},
		{"Queued", "queued"},
		{"Stopped", "canceled"},
		{"Canceled", "canceled"},
		{"Unknown", "unknown"},
	}

	for _, tt := range tests {
		result := normalizeStatus(tt.input)
		if result != tt.expected {
			t.Errorf("normalizeStatus(%s) = %s; want %s", tt.input, result, tt.expected)
		}
	}
}

func TestNormalizeFindingStatus(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Open", "open"},
		{"New", "open"},
		{"Active", "open"},
		{"Fixed", "fixed"},
		{"Closed", "fixed"},
		{"Ignored", "ignored"},
		{"False Positive", "ignored"},
		{"Noise", "noise"},
		{"Unknown", "unknown"},
	}

	for _, tt := range tests {
		result := normalizeFindingStatus(tt.input)
		if result != tt.expected {
			t.Errorf("normalizeFindingStatus(%s) = %s; want %s", tt.input, result, tt.expected)
		}
	}
}

func TestNormalizeReportStatus(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Ready", "ready"},
		{"Completed", "ready"},
		{"In Progress", "pending"},
		{"Running", "pending"},
		{"Pending", "pending"},
		{"Failed", "failed"},
		{"Error", "failed"},
		{"Unknown", "unknown"},
	}

	for _, tt := range tests {
		result := normalizeReportStatus(tt.input)
		if result != tt.expected {
			t.Errorf("normalizeReportStatus(%s) = %s; want %s", tt.input, result, tt.expected)
		}
	}
}

func TestApplicationNormalize(t *testing.T) {
	raw := map[string]any{
		"Id":            "app-123",
		"Name":          "Test App",
		"Description":   "A test application",
		"BusinessUnit":  "Engineering",
		"Tags":          []any{"web", "production"},
		"AssetGroupIds": []any{"ag-1", "ag-2"},
		"Created":       "2024-01-15T10:00:00Z",
		"LastModified":  "2024-01-20T15:30:00Z",
	}

	app := Application(raw)

	if app.ID != "app-123" {
		t.Errorf("ID = %s; want app-123", app.ID)
	}
	if app.Name != "Test App" {
		t.Errorf("Name = %s; want Test App", app.Name)
	}
	if app.Description == nil || *app.Description != "A test application" {
		t.Errorf("Description = %v; want A test application", app.Description)
	}
	if len(app.Tags) != 2 || app.Tags[0] != "web" {
		t.Errorf("Tags = %v; want [web, production]", app.Tags)
	}
	if len(app.AssetGroupIDs) != 2 {
		t.Errorf("AssetGroupIDs = %v; want [ag-1, ag-2]", app.AssetGroupIDs)
	}
	if app.Raw == nil {
		t.Error("Raw should not be nil")
	}
}

func TestScanNormalize(t *testing.T) {
	raw := map[string]any{
		"Id":              "scan-456",
		"ApplicationId":   "app-123",
		"ScanType":        "DAST",
		"State":           "Running",
		"ExecutionStatus": "Started",
		"Url":             "https://example.com",
		"SubmissionTime":  "2024-01-15T10:00:00Z",
		"StartTime":       "2024-01-15T10:05:00Z",
		"FinishTime":      "2024-01-15T11:00:00Z",
	}

	scan := Scan(raw)

	if scan.ID != "scan-456" {
		t.Errorf("ID = %s; want scan-456", scan.ID)
	}
	if scan.Status != "running" {
		t.Errorf("Status = %s; want running", scan.Status)
	}
	if scan.QueueState != "not_queued" {
		t.Errorf("QueueState = %s; want not_queued", scan.QueueState)
	}
	if scan.Target == nil || *scan.Target != "https://example.com" {
		t.Errorf("Target = %v; want https://example.com", scan.Target)
	}
}

func TestFindingNormalize(t *testing.T) {
	raw := map[string]any{
		"Id":            "finding-789",
		"ApplicationId": "app-123",
		"ScanId":        "scan-456",
		"IssueName":     "SQL Injection",
		"Severity":      "High",
		"Status":        "Open",
		"IssueType":     "Vulnerability",
		"Location":      "https://example.com/login",
	}

	finding := Finding(raw)

	if finding.ID != "finding-789" {
		t.Errorf("ID = %s; want finding-789", finding.ID)
	}
	if finding.Title != "SQL Injection" {
		t.Errorf("Title = %s; want SQL Injection", finding.Title)
	}
	if finding.Severity != "high" {
		t.Errorf("Severity = %s; want high", finding.Severity)
	}
	if finding.Status != "open" {
		t.Errorf("Status = %s; want open", finding.Status)
	}
}

func TestReportNormalize(t *testing.T) {
	raw := map[string]any{
		"Id":            "report-101",
		"ApplicationId": "app-123",
		"ScanId":        "scan-456",
		"Status":        "Ready",
		"ReportType":    "PDF",
		"DownloadUrl":   "https://cloud.appscan.com/reports/download/101",
		"CreatedDate":   "2024-01-15T12:00:00Z",
	}

	report := Report(raw)

	if report.ID != "report-101" {
		t.Errorf("ID = %s; want report-101", report.ID)
	}
	if report.Status != "ready" {
		t.Errorf("Status = %s; want ready", report.Status)
	}
	if report.Format != "PDF" {
		t.Errorf("Format = %s; want PDF", report.Format)
	}
	if report.DownloadURL == nil || *report.DownloadURL != "https://cloud.appscan.com/reports/download/101" {
		t.Errorf("DownloadURL = %v; want https://cloud.appscan.com/reports/download/101", report.DownloadURL)
	}
}
