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
// ScansListTool tests
// ---------------------------------------------------------------------------

func TestScansListTool_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "test-key-id:test-key-secret", r.Header.Get("X-Api-Key"))
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Equal(t, "application/json", r.Header.Get("Accept"))
		assert.Equal(t, "GET", r.Method)
		// ASoC v4 uses /Scans (PascalCase)
		assert.Equal(t, "/Scans", r.URL.Path)
		// OData parameters: $skip and $top
		assert.Equal(t, "0", r.URL.Query().Get("$skip"))
		assert.Equal(t, "20", r.URL.Query().Get("$top"))

		response := map[string]interface{}{
			"Items": []interface{}{
				map[string]interface{}{
					"Id":            "scan-id-1",
					"State":         "Ready",
					"ApplicationId": "app-id-1",
					"ScanType":      "DAST",
				},
				map[string]interface{}{
					"Id":            "scan-id-2",
					"State":         "Running",
					"ApplicationId": "app-id-2",
					"ScanType":      "DAST",
				},
			},
			"TotalCount": 2,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer ts.Close()

	c := client.New(ts.URL, "test-key-id", "test-key-secret", 30)
	tool := NewScansListTool(c, nil)

	result, err := tool.Handle(context.Background(), map[string]interface{}{})
	assert.NoError(t, err)
	assert.NotEmpty(t, result)

	var output map[string]interface{}
	err = json.Unmarshal([]byte(result), &output)
	assert.NoError(t, err)

	scans := output["scans"].([]interface{})
	assert.Len(t, scans, 2)

	first := scans[0].(map[string]interface{})
	assert.Equal(t, "scan-id-1", first["id"])
	// State "Ready" normalizes to "completed"
	assert.Equal(t, "completed", first["status"])

	second := scans[1].(map[string]interface{})
	assert.Equal(t, "scan-id-2", second["id"])
	assert.Equal(t, "running", second["status"])

	assert.Equal(t, float64(1), output["page"])
	assert.Equal(t, float64(20), output["page_size"])
	assert.Equal(t, float64(2), output["total_count"])
}

func TestScansListTool_WithFilters(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Debug: log the request
		t.Logf("Request: %s %s?%s", r.Method, r.URL.Path, r.URL.RawQuery)

		// Just verify basic request structure
		if r.Method != "GET" {
			t.Logf("Method mismatch: expected GET, got %s", r.Method)
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		if r.URL.Path != "/Scans" {
			t.Logf("Path mismatch: expected /Scans, got %s", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
			return
		}

		response := map[string]interface{}{
			"Items":      []interface{}{},
			"TotalCount": 0,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer ts.Close()

	c := client.New(ts.URL, "test-key-id", "test-key-secret", 30)
	tool := NewScansListTool(c, nil)

	result, err := tool.Handle(context.Background(), map[string]interface{}{
		"page":           float64(2),
		"page_size":      float64(10),
		"application_id": "app-id-42",
		"scan_family":    "DAST",
		"status":         "running",
	})
	if err != nil {
		t.Fatalf("Handle error: %v", err)
	}

	var output map[string]interface{}
	err = json.Unmarshal([]byte(result), &output)
	if err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}
	assert.Equal(t, float64(2), output["page"])
	assert.Equal(t, float64(10), output["page_size"])
}

// ---------------------------------------------------------------------------
// ScanGetTool tests
// ---------------------------------------------------------------------------

func TestScanGetTool_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "test-key-id:test-key-secret", r.Header.Get("X-Api-Key"))
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Equal(t, "application/json", r.Header.Get("Accept"))
		assert.Equal(t, "GET", r.Method)
		// Now uses list endpoint with $top filter
		assert.Equal(t, "/Scans", r.URL.Path)
		assert.Equal(t, "500", r.URL.Query().Get("$top"))

		// Return list containing the scan we want
		response := map[string]interface{}{
			"Items": []interface{}{
				map[string]interface{}{
					"Id":              "scan-id-1",
					"ApplicationId":   "app-id-1",
					"State":           "Ready",
					"ExecutionStatus": "Started",
					"ScanType":        "DAST",
					"Url":             "https://example.com",
				},
			},
			"TotalCount": 1,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer ts.Close()

	c := client.New(ts.URL, "test-key-id", "test-key-secret", 30)
	tool := NewScanGetTool(c, nil)

	result, err := tool.Handle(context.Background(), map[string]interface{}{"id": "scan-id-1"})
	assert.NoError(t, err)
	assert.NotEmpty(t, result)

	var output map[string]interface{}
	err = json.Unmarshal([]byte(result), &output)
	assert.NoError(t, err)

	assert.Equal(t, "scan-id-1", output["id"])
	assert.Equal(t, "completed", output["status"])       // "Ready" → "completed"
	assert.Equal(t, "not_queued", output["queue_state"]) // "Started" → "not_queued"
	assert.Equal(t, "DAST", output["scan_family"])

	// app_id is a pointer, serializes as string
	assert.Equal(t, "app-id-1", output["app_id"])

	// raw field preserved
	raw := output["raw"].(map[string]interface{})
	assert.Equal(t, "scan-id-1", raw["Id"])
}

func TestScanGetTool_NotFound(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		// Now uses list endpoint
		assert.Equal(t, "/Scans", r.URL.Path)

		// Return list without the scan we're looking for
		response := map[string]interface{}{
			"Items": []interface{}{
				map[string]interface{}{
					"Id":   "other-scan-id",
					"Name": "Other Scan",
				},
			},
			"TotalCount": 1,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer ts.Close()

	c := client.New(ts.URL, "test-key-id", "test-key-secret", 30)
	tool := NewScanGetTool(c, nil)

	_, err := tool.Handle(context.Background(), map[string]interface{}{"id": "no-such-scan"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "scan not found: no-such-scan")
}

func TestScanGetTool_MissingID(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("request should not be made when id is missing")
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	c := client.New(ts.URL, "test-key-id", "test-key-secret", 30)
	tool := NewScanGetTool(c, nil)

	_, err := tool.Handle(context.Background(), map[string]interface{}{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "id is required")
}

// ---------------------------------------------------------------------------
// ScanStatusTool tests
// ---------------------------------------------------------------------------

func TestScanStatusTool_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "test-key-id:test-key-secret", r.Header.Get("X-Api-Key"))
		assert.Equal(t, "GET", r.Method)
		// Now uses list endpoint with $top filter
		assert.Equal(t, "/Scans", r.URL.Path)
		assert.Equal(t, "500", r.URL.Query().Get("$top"))

		// Return list containing the scan we want
		response := map[string]interface{}{
			"Items": []interface{}{
				map[string]interface{}{
					"Id":              "scan-id-99",
					"State":           "Running",
					"ExecutionStatus": "Queued",
					"SubmissionTime":  "2024-01-15T10:00:00Z",
					"StartTime":       "2024-01-15T10:05:00Z",
				},
			},
			"TotalCount": 1,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer ts.Close()

	c := client.New(ts.URL, "test-key-id", "test-key-secret", 30)
	tool := NewScanStatusTool(c, nil)

	result, err := tool.Handle(context.Background(), map[string]interface{}{"id": "scan-id-99"})
	assert.NoError(t, err)
	assert.NotEmpty(t, result)

	var output map[string]interface{}
	err = json.Unmarshal([]byte(result), &output)
	assert.NoError(t, err)

	// Verify the status-focused fields are present
	assert.Equal(t, "scan-id-99", output["id"])
	assert.Equal(t, "running", output["status"])     // "Running" → "running"
	assert.Equal(t, "queued", output["queue_state"]) // "Queued" → "queued"

	// Timestamps present (non-null)
	assert.NotNil(t, output["submitted_at"])
	assert.NotNil(t, output["started_at"])

	// completed_at is nil for a running scan - should be JSON null
	_, hasCompletedAt := output["completed_at"]
	assert.True(t, hasCompletedAt, "completed_at key should be present")
}

func TestScanStatusTool_MissingID(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("request should not be made when id is missing")
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	c := client.New(ts.URL, "test-key-id", "test-key-secret", 30)
	tool := NewScanStatusTool(c, nil)

	_, err := tool.Handle(context.Background(), map[string]interface{}{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "id is required")
}

// ---------------------------------------------------------------------------
// ScanCancelTool tests
// ---------------------------------------------------------------------------

func TestScanCancelTool_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "test-key-id:test-key-secret", r.Header.Get("X-Api-Key"))
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/scans/scan-id-77/cancel", r.URL.Path)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	c := client.New(ts.URL, "test-key-id", "test-key-secret", 30)
	tool := NewScanCancelTool(c, nil)

	result, err := tool.Handle(context.Background(), map[string]interface{}{"id": "scan-id-77"})
	assert.NoError(t, err)
	assert.NotEmpty(t, result)

	var output map[string]interface{}
	err = json.Unmarshal([]byte(result), &output)
	assert.NoError(t, err)

	assert.Equal(t, "scan-id-77", output["id"])
	assert.Equal(t, "Scan cancellation requested", output["message"])
}

func TestScanCancelTool_NoContent(t *testing.T) {
	// Some ASoC tenants may return 204 No Content on cancel
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/scans/scan-id-88/cancel", r.URL.Path)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer ts.Close()

	c := client.New(ts.URL, "test-key-id", "test-key-secret", 30)
	tool := NewScanCancelTool(c, nil)

	result, err := tool.Handle(context.Background(), map[string]interface{}{"id": "scan-id-88"})
	assert.NoError(t, err)

	var output map[string]interface{}
	err = json.Unmarshal([]byte(result), &output)
	assert.NoError(t, err)
	assert.Equal(t, "scan-id-88", output["id"])
	assert.Equal(t, "Scan cancellation requested", output["message"])
}

func TestScanCancelTool_MissingID(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("request should not be made when id is missing")
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	c := client.New(ts.URL, "test-key-id", "test-key-secret", 30)
	tool := NewScanCancelTool(c, nil)

	_, err := tool.Handle(context.Background(), map[string]interface{}{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "id is required")
}

func TestScanCancelTool_ReadOnly(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("request should not be made in readonly mode")
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	c := client.New(ts.URL, "test-key-id", "test-key-secret", 30)
	tool := NewScanCancelTool(c, &readonlyFlag{true})

	_, err := tool.Handle(context.Background(), map[string]interface{}{"id": "scan-id-77"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "readonly mode")
}

func TestScanCancelTool_ServerError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	c := client.New(ts.URL, "test-key-id", "test-key-secret", 30)
	tool := NewScanCancelTool(c, nil)

	_, err := tool.Handle(context.Background(), map[string]interface{}{"id": "scan-id-77"})
	assert.Error(t, err)
}
