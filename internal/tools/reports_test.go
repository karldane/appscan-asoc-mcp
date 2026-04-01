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
// ReportsListTool tests
// ---------------------------------------------------------------------------

func TestReportsListTool_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "test-key-id:test-key-secret", r.Header.Get("X-Api-Key"))
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Equal(t, "application/json", r.Header.Get("Accept"))
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/reports", r.URL.Path)
		assert.Equal(t, "1", r.URL.Query().Get("page"))
		assert.Equal(t, "20", r.URL.Query().Get("pageSize"))

		response := map[string]interface{}{
			"Reports": []interface{}{
				map[string]interface{}{
					"Id":            "report-id-1",
					"ApplicationId": "app-id-1",
					"Status":        "Ready",
					"ReportType":    "PDF",
					"DownloadUrl":   "https://example.com/reports/report-id-1.pdf",
				},
			},
			"TotalPages": 1,
			"TotalCount": 1,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer ts.Close()

	c := client.New(ts.URL, "test-key-id", "test-key-secret", 30)
	tool := NewReportsListTool(c, nil)

	result, err := tool.Handle(context.Background(), map[string]interface{}{})
	assert.NoError(t, err)
	assert.NotEmpty(t, result)

	var output map[string]interface{}
	err = json.Unmarshal([]byte(result), &output)
	assert.NoError(t, err)

	reports := output["reports"].([]interface{})
	assert.Len(t, reports, 1)

	first := reports[0].(map[string]interface{})
	assert.Equal(t, "report-id-1", first["id"])
	assert.Equal(t, "app-id-1", first["app_id"])
	// "Ready" normalizes to "ready"
	assert.Equal(t, "ready", first["status"])
	assert.Equal(t, "PDF", first["format"])
	assert.Equal(t, "https://example.com/reports/report-id-1.pdf", first["download_url"])

	// raw field preserved
	raw := first["raw"].(map[string]interface{})
	assert.Equal(t, "report-id-1", raw["Id"])

	assert.Equal(t, float64(1), output["page"])
	assert.Equal(t, float64(20), output["page_size"])
	assert.Equal(t, float64(1), output["total_pages"])
	assert.Equal(t, float64(1), output["total_count"])
}

func TestReportsListTool_ReadOnly(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("request should not be made in readonly mode")
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	c := client.New(ts.URL, "test-key-id", "test-key-secret", 30)
	tool := NewReportsListTool(c, &readonlyFlag{true})

	_, err := tool.Handle(context.Background(), map[string]interface{}{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "readonly mode")
}

// ---------------------------------------------------------------------------
// ReportGenerateTool tests
// ---------------------------------------------------------------------------

func TestReportGenerateTool_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "test-key-id:test-key-secret", r.Header.Get("X-Api-Key"))
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Equal(t, "application/json", r.Header.Get("Accept"))
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/reports/generate", r.URL.Path)

		// Decode and verify request body
		var body map[string]interface{}
		json.NewDecoder(r.Body).Decode(&body)
		assert.Equal(t, "app-id-1", body["ApplicationId"])
		assert.Equal(t, "PDF", body["ReportType"])

		response := map[string]interface{}{
			"Id":     "new-report-id-1",
			"Status": "Pending",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer ts.Close()

	c := client.New(ts.URL, "test-key-id", "test-key-secret", 30)
	tool := NewReportGenerateTool(c, nil)

	result, err := tool.Handle(context.Background(), map[string]interface{}{
		"application_id": "app-id-1",
		"report_type":    "PDF",
	})
	assert.NoError(t, err)
	assert.NotEmpty(t, result)

	var output map[string]interface{}
	err = json.Unmarshal([]byte(result), &output)
	assert.NoError(t, err)

	assert.Equal(t, "new-report-id-1", output["id"])
	assert.Equal(t, "Pending", output["status"])
	assert.Equal(t, "Report generation started", output["message"])
}

func TestReportGenerateTool_DefaultType(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/reports/generate", r.URL.Path)

		var body map[string]interface{}
		json.NewDecoder(r.Body).Decode(&body)
		// When report_type is omitted the tool defaults to "PDF"
		assert.Equal(t, "PDF", body["ReportType"])

		response := map[string]interface{}{
			"Id":     "new-report-id-2",
			"Status": "Pending",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer ts.Close()

	c := client.New(ts.URL, "test-key-id", "test-key-secret", 30)
	tool := NewReportGenerateTool(c, nil)

	result, err := tool.Handle(context.Background(), map[string]interface{}{
		"application_id": "app-id-1",
	})
	assert.NoError(t, err)

	var output map[string]interface{}
	err = json.Unmarshal([]byte(result), &output)
	assert.NoError(t, err)
	assert.Equal(t, "new-report-id-2", output["id"])
}

func TestReportGenerateTool_ReadOnly(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("request should not be made in readonly mode")
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	c := client.New(ts.URL, "test-key-id", "test-key-secret", 30)
	tool := NewReportGenerateTool(c, &readonlyFlag{true})

	_, err := tool.Handle(context.Background(), map[string]interface{}{
		"application_id": "app-id-1",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "readonly mode")
}

// ---------------------------------------------------------------------------
// ReportGetTool tests
// ---------------------------------------------------------------------------

func TestReportGetTool_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "test-key-id:test-key-secret", r.Header.Get("X-Api-Key"))
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Equal(t, "application/json", r.Header.Get("Accept"))
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/reports/report-id-1", r.URL.Path)

		response := map[string]interface{}{
			"Id":            "report-id-1",
			"ApplicationId": "app-id-1",
			"ScanId":        "scan-id-1",
			"Status":        "Ready",
			"ReportType":    "PDF",
			"DownloadUrl":   "https://example.com/reports/report-id-1.pdf",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer ts.Close()

	c := client.New(ts.URL, "test-key-id", "test-key-secret", 30)
	tool := NewReportGetTool(c, nil)

	result, err := tool.Handle(context.Background(), map[string]interface{}{"id": "report-id-1"})
	assert.NoError(t, err)
	assert.NotEmpty(t, result)

	var output map[string]interface{}
	err = json.Unmarshal([]byte(result), &output)
	assert.NoError(t, err)

	assert.Equal(t, "report-id-1", output["id"])
	assert.Equal(t, "app-id-1", output["app_id"])
	assert.Equal(t, "scan-id-1", output["scan_id"])
	// "Ready" normalizes to "ready"
	assert.Equal(t, "ready", output["status"])
	assert.Equal(t, "PDF", output["format"])
	assert.Equal(t, "https://example.com/reports/report-id-1.pdf", output["download_url"])

	// raw field preserved
	raw := output["raw"].(map[string]interface{})
	assert.Equal(t, "report-id-1", raw["Id"])
}

func TestReportGetTool_NotFound(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/reports/no-such-report", r.URL.Path)
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	c := client.New(ts.URL, "test-key-id", "test-key-secret", 30)
	tool := NewReportGetTool(c, nil)

	_, err := tool.Handle(context.Background(), map[string]interface{}{"id": "no-such-report"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "report not found: no-such-report")
}

func TestReportGetTool_ReadOnly(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("request should not be made in readonly mode")
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	c := client.New(ts.URL, "test-key-id", "test-key-secret", 30)
	tool := NewReportGetTool(c, &readonlyFlag{true})

	_, err := tool.Handle(context.Background(), map[string]interface{}{"id": "report-id-1"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "readonly mode")
}
