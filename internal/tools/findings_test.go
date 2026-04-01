package tools

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/karldane/appscan-asoc-mcp/internal/client"
	"github.com/stretchr/testify/assert"
)

// ---------------------------------------------------------------------------
// FindingsListTool tests
// ---------------------------------------------------------------------------

func TestFindingsListTool_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "test-key-id:test-key-secret", r.Header.Get("X-Api-Key"))
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Equal(t, "application/json", r.Header.Get("Accept"))
		assert.Equal(t, "GET", r.Method)
		// New endpoint: /Issues/Application/{appId}
		assert.Equal(t, "/Issues/Application/app-id-1", r.URL.Path)
		assert.Equal(t, "0", r.URL.Query().Get("$skip"))
		assert.Equal(t, "50", r.URL.Query().Get("$top"))

		response := map[string]interface{}{
			"Items": []interface{}{
				map[string]interface{}{
					"Id":            "finding-id-1",
					"IssueName":     "SQL Injection",
					"Severity":      "High",
					"Status":        "Open",
					"ApplicationId": "app-id-1",
					"ScanId":        "scan-id-1",
				},
				map[string]interface{}{
					"Id":            "finding-id-2",
					"IssueName":     "XSS",
					"Severity":      "Medium",
					"Status":        "Fixed",
					"ApplicationId": "app-id-1",
					"ScanId":        "scan-id-1",
				},
			},
			"TotalCount": 2,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer ts.Close()

	c := client.New(ts.URL, "test-key-id", "test-key-secret", 30)
	tool := NewFindingsListTool(c, nil)

	// Now application_id is required
	result, err := tool.Handle(context.Background(), map[string]interface{}{
		"application_id": "app-id-1",
	})
	assert.NoError(t, err)
	assert.NotEmpty(t, result)

	var output map[string]interface{}
	err = json.Unmarshal([]byte(result), &output)
	assert.NoError(t, err)

	findings := output["findings"].([]interface{})
	assert.Len(t, findings, 2)

	first := findings[0].(map[string]interface{})
	assert.Equal(t, "finding-id-1", first["id"])
	assert.Equal(t, "SQL Injection", first["title"])
	assert.Equal(t, "high", first["severity"])
	assert.Equal(t, "open", first["status"])

	second := findings[1].(map[string]interface{})
	assert.Equal(t, "finding-id-2", second["id"])
	assert.Equal(t, "XSS", second["title"])
	assert.Equal(t, "medium", second["severity"])
	assert.Equal(t, "fixed", second["status"])

	assert.Equal(t, float64(1), output["page"])
	assert.Equal(t, float64(50), output["page_size"])
	assert.Equal(t, float64(2), output["total_count"])
}

func TestFindingsListTool_WithFilters(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Return a successful response
		response := map[string]interface{}{
			"Items":      []interface{}{},
			"TotalCount": 0,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer ts.Close()

	c := client.New(ts.URL, "test-key-id", "test-key-secret", 30)
	tool := NewFindingsListTool(c, nil)

	// Test without scan_id filter first
	result, err := tool.Handle(context.Background(), map[string]interface{}{
		"page":           float64(2),
		"page_size":      float64(10),
		"application_id": "app-id-42",
	})
	if err != nil {
		t.Fatalf("Handle without scan_id returned error: %v", err)
	}

	var output map[string]interface{}
	err = json.Unmarshal([]byte(result), &output)
	if err != nil {
		t.Fatalf("Failed to unmarshal result: %v", err)
	}
	assert.Equal(t, float64(2), output["page"])
	assert.Equal(t, float64(10), output["page_size"])

	// Test with scan_id filter
	result2, err := tool.Handle(context.Background(), map[string]interface{}{
		"page":           float64(2),
		"page_size":      float64(10),
		"application_id": "app-id-42",
		"scan_id":        "scan-id-99",
	})
	if err != nil {
		t.Fatalf("Handle with scan_id returned error: %v", err)
	}

	var output2 map[string]interface{}
	err = json.Unmarshal([]byte(result2), &output2)
	if err != nil {
		t.Fatalf("Failed to unmarshal result: %v", err)
	}
	assert.Equal(t, float64(2), output2["page"])
	assert.Equal(t, float64(10), output2["page_size"])
}

// ---------------------------------------------------------------------------
// FindingsSearchTool tests
// ---------------------------------------------------------------------------

func TestFindingsSearchTool_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "test-key-id:test-key-secret", r.Header.Get("X-Api-Key"))
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Equal(t, "application/json", r.Header.Get("Accept"))
		// Now uses GET with OData parameters
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/Issues/Application/app-id-1", r.URL.Path)
		q := r.URL.Query()
		assert.Equal(t, "0", q.Get("$skip"))
		assert.Equal(t, "50", q.Get("$top"))
		filter := q.Get("$filter")
		// URL-encoded filter values
		assert.Contains(t, filter, "High")
		assert.Contains(t, filter, "Open")

		response := map[string]interface{}{
			"Items": []interface{}{
				map[string]interface{}{
					"Id":            "finding-id-1",
					"IssueName":     "SQL Injection",
					"Severity":      "High",
					"Status":        "Open",
					"ApplicationId": "app-id-1",
				},
			},
			"TotalCount": 1,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer ts.Close()

	c := client.New(ts.URL, "test-key-id", "test-key-secret", 30)
	tool := NewFindingsSearchTool(c, nil)

	result, err := tool.Handle(context.Background(), map[string]interface{}{
		"application_id": "app-id-1",
		"severity":       "High",
		"status":         "Open",
	})
	assert.NoError(t, err)
	assert.NotEmpty(t, result)

	var output map[string]interface{}
	err = json.Unmarshal([]byte(result), &output)
	assert.NoError(t, err)

	findings := output["findings"].([]interface{})
	assert.Len(t, findings, 1)

	first := findings[0].(map[string]interface{})
	assert.Equal(t, "finding-id-1", first["id"])
	assert.Equal(t, "SQL Injection", first["title"])
	assert.Equal(t, "high", first["severity"])
	assert.Equal(t, "open", first["status"])

	assert.Equal(t, float64(1), output["total_count"])
}

// ---------------------------------------------------------------------------
// FindingGetTool tests
// ---------------------------------------------------------------------------

func TestFindingGetTool_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "test-key-id:test-key-secret", r.Header.Get("X-Api-Key"))
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Equal(t, "application/json", r.Header.Get("Accept"))
		assert.Equal(t, "GET", r.Method)
		// ASoC v4 endpoint: /Issues/{id} (not /Issues/{id}/Details which returns HTML)
		assert.Equal(t, "/Issues/finding-id-1", r.URL.Path)

		response := map[string]interface{}{
			"Id":            "finding-id-1",
			"IssueName":     "SQL Injection",
			"Severity":      "Critical",
			"Status":        "Open",
			"ApplicationId": "app-id-1",
			"ScanId":        "scan-id-1",
			"IssueType":     "Injection",
			"Location":      "https://example.com/login",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer ts.Close()

	c := client.New(ts.URL, "test-key-id", "test-key-secret", 30)
	tool := NewFindingGetTool(c, nil)

	result, err := tool.Handle(context.Background(), map[string]interface{}{"id": "finding-id-1"})
	assert.NoError(t, err)
	assert.NotEmpty(t, result)

	var output map[string]interface{}
	err = json.Unmarshal([]byte(result), &output)
	assert.NoError(t, err)

	assert.Equal(t, "finding-id-1", output["id"])
	assert.Equal(t, "SQL Injection", output["title"])
	// "Critical" normalizes to "high"
	assert.Equal(t, "high", output["severity"])
	assert.Equal(t, "open", output["status"])
	assert.Equal(t, "Injection", output["issue_type"])
	assert.Equal(t, "https://example.com/login", output["location"])

	// raw field preserved
	raw := output["raw"].(map[string]interface{})
	assert.Equal(t, "finding-id-1", raw["Id"])
}

func TestFindingGetTool_NotFound(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		// ASoC v4 endpoint: /Issues/{id} (not /Issues/{id}/Details)
		assert.Equal(t, "/Issues/no-such-finding", r.URL.Path)
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	c := client.New(ts.URL, "test-key-id", "test-key-secret", 30)
	tool := NewFindingGetTool(c, nil)

	_, err := tool.Handle(context.Background(), map[string]interface{}{"id": "no-such-finding"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "finding not found: no-such-finding")
}

// ---------------------------------------------------------------------------
// FindingGroupSummaryTool tests
// ---------------------------------------------------------------------------

func TestFindingGroupSummaryTool_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "test-key-id:test-key-secret", r.Header.Get("X-Api-Key"))
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Equal(t, "application/json", r.Header.Get("Accept"))
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/findings/summary/severity", r.URL.Path)

		// Verify request body contains the application ID
		var body map[string]interface{}
		json.NewDecoder(r.Body).Decode(&body)
		assert.Equal(t, "app-id-1", body["ApplicationId"])

		response := map[string]interface{}{
			"Groups": []interface{}{
				map[string]interface{}{
					"Key":   "High",
					"Count": 5,
				},
				map[string]interface{}{
					"Key":   "Medium",
					"Count": 12,
				},
			},
			"TotalCount": 17,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer ts.Close()

	c := client.New(ts.URL, "test-key-id", "test-key-secret", 30)
	tool := NewFindingGroupSummaryTool(c, nil)

	result, err := tool.Handle(context.Background(), map[string]interface{}{
		"application_id": "app-id-1",
		"group_by":       "severity",
	})
	assert.NoError(t, err)
	assert.NotEmpty(t, result)

	var output map[string]interface{}
	err = json.Unmarshal([]byte(result), &output)
	assert.NoError(t, err)

	assert.NotNil(t, output["Groups"])
	assert.Equal(t, float64(17), output["TotalCount"])
}

func TestFindingGroupSummaryTool_DefaultGroupBy(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// When group_by is omitted, tool should default to "severity"
		assert.Equal(t, "/findings/summary/severity", r.URL.Path)

		response := map[string]interface{}{"Groups": []interface{}{}}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer ts.Close()

	c := client.New(ts.URL, "test-key-id", "test-key-secret", 30)
	tool := NewFindingGroupSummaryTool(c, nil)

	result, err := tool.Handle(context.Background(), map[string]interface{}{
		"application_id": "app-id-1",
	})
	assert.NoError(t, err)
	assert.NotEmpty(t, result)
}
