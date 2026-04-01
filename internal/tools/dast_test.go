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
// DASTScanStartTool tests
// ---------------------------------------------------------------------------

func TestDASTScanStartTool_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "test-key-id:test-key-secret", r.Header.Get("X-Api-Key"))
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Equal(t, "application/json", r.Header.Get("Accept"))
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/scans/dast", r.URL.Path)

		var body map[string]interface{}
		json.NewDecoder(r.Body).Decode(&body)
		assert.Equal(t, "app-id-1", body["ApplicationId"])
		assert.Equal(t, "https://example.com", body["Url"])

		response := map[string]interface{}{
			"Id":              "scan-id-1",
			"State":           "Queued",
			"ExecutionStatus": "Queued",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer ts.Close()

	c := client.New(ts.URL, "test-key-id", "test-key-secret", 30)
	tool := NewDASTScanStartTool(c, nil)

	result, err := tool.Handle(context.Background(), map[string]interface{}{
		"application_id": "app-id-1",
		"target_url":     "https://example.com",
	})
	assert.NoError(t, err)
	assert.NotEmpty(t, result)

	var output map[string]interface{}
	err = json.Unmarshal([]byte(result), &output)
	assert.NoError(t, err)

	assert.Equal(t, "scan-id-1", output["id"])
	// "Queued" State → normalizeStatus → "queued"
	assert.Equal(t, "queued", output["status"])
	// "Queued" ExecutionStatus → normalizeQueueState → "queued"
	assert.Equal(t, "queued", output["queue_state"])
	assert.Equal(t, "Scan started or queued", output["message"])
}

func TestDASTScanStartTool_Created(t *testing.T) {
	// Some ASoC tenants return 201 Created
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/scans/dast", r.URL.Path)

		response := map[string]interface{}{
			"Id":              "scan-id-2",
			"State":           "Running",
			"ExecutionStatus": "Started",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(response)
	}))
	defer ts.Close()

	c := client.New(ts.URL, "test-key-id", "test-key-secret", 30)
	tool := NewDASTScanStartTool(c, nil)

	result, err := tool.Handle(context.Background(), map[string]interface{}{
		"application_id": "app-id-2",
		"target_url":     "https://example.com/app2",
	})
	assert.NoError(t, err)

	var output map[string]interface{}
	err = json.Unmarshal([]byte(result), &output)
	assert.NoError(t, err)

	assert.Equal(t, "scan-id-2", output["id"])
	assert.Equal(t, "running", output["status"])
	assert.Equal(t, "not_queued", output["queue_state"])
}

func TestDASTScanStartTool_MissingArgs(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("request should not be made when required args are missing")
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	c := client.New(ts.URL, "test-key-id", "test-key-secret", 30)
	tool := NewDASTScanStartTool(c, nil)

	// Missing both
	_, err := tool.Handle(context.Background(), map[string]interface{}{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "application_id and target_url are required")

	// Missing target_url
	_, err = tool.Handle(context.Background(), map[string]interface{}{
		"application_id": "app-id-1",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "application_id and target_url are required")

	// Missing application_id
	_, err = tool.Handle(context.Background(), map[string]interface{}{
		"target_url": "https://example.com",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "application_id and target_url are required")
}

func TestDASTScanStartTool_ReadOnly(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("request should not be made in readonly mode")
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	c := client.New(ts.URL, "test-key-id", "test-key-secret", 30)
	tool := NewDASTScanStartTool(c, &readonlyFlag{true})

	_, err := tool.Handle(context.Background(), map[string]interface{}{
		"application_id": "app-id-1",
		"target_url":     "https://example.com",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "readonly mode")
}

func TestDASTScanStartTool_ServerError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	c := client.New(ts.URL, "test-key-id", "test-key-secret", 30)
	tool := NewDASTScanStartTool(c, nil)

	_, err := tool.Handle(context.Background(), map[string]interface{}{
		"application_id": "app-id-1",
		"target_url":     "https://example.com",
	})
	assert.Error(t, err)
}

// ---------------------------------------------------------------------------
// DASTScanFromTemplateTool tests
// ---------------------------------------------------------------------------

func TestDASTScanFromTemplateTool_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "test-key-id:test-key-secret", r.Header.Get("X-Api-Key"))
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Equal(t, "application/json", r.Header.Get("Accept"))
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/scans/dast/fromfile", r.URL.Path)

		var body map[string]interface{}
		json.NewDecoder(r.Body).Decode(&body)
		assert.Equal(t, "app-id-1", body["ApplicationId"])
		assert.Equal(t, "file-id-1", body["ScanOrTemplateFileId"])

		response := map[string]interface{}{
			"Id":              "scan-id-1",
			"State":           "Queued",
			"ExecutionStatus": "Queued",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer ts.Close()

	c := client.New(ts.URL, "test-key-id", "test-key-secret", 30)
	tool := NewDASTScanFromTemplateTool(c, nil)

	result, err := tool.Handle(context.Background(), map[string]interface{}{
		"application_id": "app-id-1",
		"file_id":        "file-id-1",
	})
	assert.NoError(t, err)
	assert.NotEmpty(t, result)

	var output map[string]interface{}
	err = json.Unmarshal([]byte(result), &output)
	assert.NoError(t, err)

	assert.Equal(t, "scan-id-1", output["id"])
	assert.Equal(t, "queued", output["status"])
	assert.Equal(t, "queued", output["queue_state"])
	assert.Equal(t, "Scan started from template", output["message"])
}

func TestDASTScanFromTemplateTool_Created(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/scans/dast/fromfile", r.URL.Path)

		response := map[string]interface{}{
			"Id":              "scan-id-3",
			"State":           "Running",
			"ExecutionStatus": "Started",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(response)
	}))
	defer ts.Close()

	c := client.New(ts.URL, "test-key-id", "test-key-secret", 30)
	tool := NewDASTScanFromTemplateTool(c, nil)

	result, err := tool.Handle(context.Background(), map[string]interface{}{
		"application_id": "app-id-3",
		"file_id":        "file-id-3",
		"scan_name":      "My Template Scan",
	})
	assert.NoError(t, err)

	var output map[string]interface{}
	err = json.Unmarshal([]byte(result), &output)
	assert.NoError(t, err)

	assert.Equal(t, "scan-id-3", output["id"])
	assert.Equal(t, "running", output["status"])
	assert.Equal(t, "not_queued", output["queue_state"])
	assert.Equal(t, "Scan started from template", output["message"])
}

func TestDASTScanFromTemplateTool_MissingArgs(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("request should not be made when required args are missing")
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	c := client.New(ts.URL, "test-key-id", "test-key-secret", 30)
	tool := NewDASTScanFromTemplateTool(c, nil)

	// Missing both
	_, err := tool.Handle(context.Background(), map[string]interface{}{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "application_id and file_id are required")

	// Missing file_id
	_, err = tool.Handle(context.Background(), map[string]interface{}{
		"application_id": "app-id-1",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "application_id and file_id are required")

	// Missing application_id
	_, err = tool.Handle(context.Background(), map[string]interface{}{
		"file_id": "file-id-1",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "application_id and file_id are required")
}

func TestDASTScanFromTemplateTool_ReadOnly(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("request should not be made in readonly mode")
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	c := client.New(ts.URL, "test-key-id", "test-key-secret", 30)
	tool := NewDASTScanFromTemplateTool(c, &readonlyFlag{true})

	_, err := tool.Handle(context.Background(), map[string]interface{}{
		"application_id": "app-id-1",
		"file_id":        "file-id-1",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "readonly mode")
}

func TestDASTScanFromTemplateTool_ServerError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	c := client.New(ts.URL, "test-key-id", "test-key-secret", 30)
	tool := NewDASTScanFromTemplateTool(c, nil)

	_, err := tool.Handle(context.Background(), map[string]interface{}{
		"application_id": "app-id-1",
		"file_id":        "file-id-1",
	})
	assert.Error(t, err)
}
