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
		assert.Equal(t, "/scans", r.URL.Path)
		assert.Equal(t, "page=1&pageSize=20", r.URL.RawQuery)

		response := map[string]interface{}{
			"Scans": []interface{}{
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
			"TotalPages": 1,
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
	assert.Equal(t, float64(1), output["total_pages"])
	assert.Equal(t, float64(2), output["total_count"])
}

func TestScansListTool_WithFilters(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/scans", r.URL.Path)
		q := r.URL.Query()
		assert.Equal(t, "2", q.Get("page"))
		assert.Equal(t, "10", q.Get("pageSize"))
		assert.Equal(t, "app-id-42", q.Get("applicationId"))
		assert.Equal(t, "DAST", q.Get("scanType"))
		assert.Equal(t, "running", q.Get("state"))

		response := map[string]interface{}{
			"Scans":      []interface{}{},
			"TotalPages": 0,
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
	assert.NoError(t, err)

	var output map[string]interface{}
	err = json.Unmarshal([]byte(result), &output)
	assert.NoError(t, err)
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
		assert.Equal(t, "/scans/scan-id-1", r.URL.Path)

		response := map[string]interface{}{
			"Id":              "scan-id-1",
			"ApplicationId":   "app-id-1",
			"State":           "Ready",
			"ExecutionStatus": "Started",
			"ScanType":        "DAST",
			"Url":             "https://example.com",
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
		assert.Equal(t, "/scans/no-such-scan", r.URL.Path)
		w.WriteHeader(http.StatusNotFound)
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
		assert.Equal(t, "/scans/scan-id-99", r.URL.Path)

		response := map[string]interface{}{
			"Id":              "scan-id-99",
			"State":           "Running",
			"ExecutionStatus": "Queued",
			"SubmissionTime":  "2024-01-15T10:00:00Z",
			"StartTime":       "2024-01-15T10:05:00Z",
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
