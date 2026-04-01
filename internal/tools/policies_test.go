package tools

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/karldane/appscan-asoc-mcp/internal/client"
	"github.com/stretchr/testify/assert"
)

// ---------------------------------------------------------------------------
// AssetGroupsListTool tests
// ---------------------------------------------------------------------------

func TestAssetGroupsListTool_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "test-key-id:test-key-secret", r.Header.Get("X-Api-Key"))
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Equal(t, "application/json", r.Header.Get("Accept"))
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/assetgroups", r.URL.Path)
		assert.Equal(t, "1", r.URL.Query().Get("page"))
		assert.Equal(t, "20", r.URL.Query().Get("pageSize"))

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"AssetGroups": []interface{}{
				map[string]interface{}{
					"Id":   "ag-1",
					"Name": "Default",
				},
			},
			"TotalPages": 1,
			"TotalCount": 1,
		})
	}))
	defer ts.Close()

	c := client.New(ts.URL, "test-key-id", "test-key-secret", 30)
	tool := NewAssetGroupsListTool(c, nil)

	result, err := tool.Handle(context.Background(), map[string]interface{}{})
	assert.NoError(t, err)
	assert.NotEmpty(t, result)

	var output map[string]interface{}
	err = json.Unmarshal([]byte(result), &output)
	assert.NoError(t, err)

	groups := output["asset_groups"].([]interface{})
	assert.Len(t, groups, 1)

	first := groups[0].(map[string]interface{})
	assert.Equal(t, "ag-1", first["Id"])
	assert.Equal(t, "Default", first["Name"])

	assert.Equal(t, float64(1), output["page"])
	assert.Equal(t, float64(20), output["page_size"])
	assert.Equal(t, float64(1), output["total_pages"])
	assert.Equal(t, float64(1), output["total_count"])
}


// ---------------------------------------------------------------------------
// PoliciesListTool tests
// ---------------------------------------------------------------------------

func TestPoliciesListTool_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "test-key-id:test-key-secret", r.Header.Get("X-Api-Key"))
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Equal(t, "application/json", r.Header.Get("Accept"))
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/policies", r.URL.Path)
		assert.Equal(t, "1", r.URL.Query().Get("page"))
		assert.Equal(t, "20", r.URL.Query().Get("pageSize"))

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"Policies": []interface{}{
				map[string]interface{}{
					"Id":   "pol-1",
					"Name": "Default Policy",
				},
			},
			"TotalPages": 1,
			"TotalCount": 1,
		})
	}))
	defer ts.Close()

	c := client.New(ts.URL, "test-key-id", "test-key-secret", 30)
	tool := NewPoliciesListTool(c, nil)

	result, err := tool.Handle(context.Background(), map[string]interface{}{})
	assert.NoError(t, err)
	assert.NotEmpty(t, result)

	var output map[string]interface{}
	err = json.Unmarshal([]byte(result), &output)
	assert.NoError(t, err)

	policies := output["policies"].([]interface{})
	assert.Len(t, policies, 1)

	first := policies[0].(map[string]interface{})
	assert.Equal(t, "pol-1", first["Id"])
	assert.Equal(t, "Default Policy", first["Name"])

	assert.Equal(t, float64(1), output["page"])
	assert.Equal(t, float64(20), output["page_size"])
	assert.Equal(t, float64(1), output["total_pages"])
	assert.Equal(t, float64(1), output["total_count"])
}


// ---------------------------------------------------------------------------
// ComplianceSummaryTool tests
// ---------------------------------------------------------------------------

func TestComplianceSummaryTool_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "test-key-id:test-key-secret", r.Header.Get("X-Api-Key"))
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Equal(t, "application/json", r.Header.Get("Accept"))
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/compliance/summary", r.URL.Path)

		// Verify request body contains ApplicationId
		body, err := io.ReadAll(r.Body)
		assert.NoError(t, err)
		var reqBody map[string]interface{}
		err = json.Unmarshal(body, &reqBody)
		assert.NoError(t, err)
		assert.Equal(t, "app-id-1", reqBody["ApplicationId"])

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"ApplicationId": "app-id-1",
			"PolicyName":    "Default Policy",
			"Status":        "Pass",
			"IssueCount":    0,
		})
	}))
	defer ts.Close()

	c := client.New(ts.URL, "test-key-id", "test-key-secret", 30)
	tool := NewComplianceSummaryTool(c, nil)

	result, err := tool.Handle(context.Background(), map[string]interface{}{
		"application_id": "app-id-1",
	})
	assert.NoError(t, err)
	assert.NotEmpty(t, result)

	var output map[string]interface{}
	err = json.Unmarshal([]byte(result), &output)
	assert.NoError(t, err)

	// Response is raw JSON passthrough
	assert.Equal(t, "app-id-1", output["ApplicationId"])
	assert.Equal(t, "Default Policy", output["PolicyName"])
	assert.Equal(t, "Pass", output["Status"])
}

func TestComplianceSummaryTool_WithScanID(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		assert.NoError(t, err)
		var reqBody map[string]interface{}
		err = json.Unmarshal(body, &reqBody)
		assert.NoError(t, err)
		assert.Equal(t, "app-id-1", reqBody["ApplicationId"])
		assert.Equal(t, "scan-id-1", reqBody["ScanId"])

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"ApplicationId": "app-id-1",
			"ScanId":        "scan-id-1",
			"Status":        "Fail",
		})
	}))
	defer ts.Close()

	c := client.New(ts.URL, "test-key-id", "test-key-secret", 30)
	tool := NewComplianceSummaryTool(c, nil)

	result, err := tool.Handle(context.Background(), map[string]interface{}{
		"application_id": "app-id-1",
		"scan_id":        "scan-id-1",
	})
	assert.NoError(t, err)
	assert.NotEmpty(t, result)

	var output map[string]interface{}
	err = json.Unmarshal([]byte(result), &output)
	assert.NoError(t, err)

	assert.Equal(t, "scan-id-1", output["ScanId"])
	assert.Equal(t, "Fail", output["Status"])
}

func TestComplianceSummaryTool_MissingApplicationID(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("request should not be made when application_id is missing")
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	c := client.New(ts.URL, "test-key-id", "test-key-secret", 30)
	tool := NewComplianceSummaryTool(c, nil)

	_, err := tool.Handle(context.Background(), map[string]interface{}{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "application_id is required")
}

