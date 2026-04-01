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

// Simple readonly flag implementation for testing
type readonlyFlag struct {
	value bool
}

func (r *readonlyFlag) ReadOnly() bool {
	return r.value
}

func TestAppsListTool_Success(t *testing.T) {
	// Create a mock HTTP server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check that the request has the correct headers
		assert.Equal(t, "test-key-id:test-key-secret", r.Header.Get("X-Api-Key"))
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Equal(t, "application/json", r.Header.Get("Accept"))
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/Apps", r.URL.Path)
		assert.Equal(t, "page=1&pageSize=20", r.URL.RawQuery)

		// Return mock response
		response := map[string]interface{}{
			"Items": []interface{}{
				map[string]interface{}{
					"Id":   "test-app-id-1",
					"Name": "Test App 1",
				},
				map[string]interface{}{
					"Id":   "test-app-id-2",
					"Name": "Test App 2",
				},
			},
			"TotalPages": 1,
			"TotalCount": 2,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer ts.Close()

	// Create client pointing to our test server
	c := client.New(ts.URL, "test-key-id", "test-key-secret", 30)

	// Create the tool with readonly disabled (nil means not readonly)
	tool := NewAppsListTool(c, nil)

	// Call the tool
	result, err := tool.Handle(context.Background(), map[string]interface{}{})
	assert.NoError(t, err)
	assert.NotEmpty(t, result)

	// Parse the result
	var output map[string]interface{}
	err = json.Unmarshal([]byte(result), &output)
	assert.NoError(t, err)

	// Check the structure
	apps := output["applications"].([]interface{})
	assert.Len(t, apps, 2)

	// Check first app
	firstApp := apps[0].(map[string]interface{})
	assert.Equal(t, "test-app-id-1", firstApp["id"])
	assert.Equal(t, "Test App 1", firstApp["name"])

	// Check second app
	secondApp := apps[1].(map[string]interface{})
	assert.Equal(t, "test-app-id-2", secondApp["id"])
	assert.Equal(t, "Test App 2", secondApp["name"])

	// Check pagination
	assert.Equal(t, float64(1), output["page"])
	assert.Equal(t, float64(20), output["page_size"])
	assert.Equal(t, float64(1), output["total_pages"])
	assert.Equal(t, float64(2), output["total_count"])
}

func TestAppGetTool_Success(t *testing.T) {
	// Create a mock HTTP server that returns a list of apps
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check that the request has the correct headers
		assert.Equal(t, "test-key-id:test-key-secret", r.Header.Get("X-Api-Key"))
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Equal(t, "application/json", r.Header.Get("Accept"))
		assert.Equal(t, "GET", r.Method)
		// New implementation uses list endpoint with query params
		assert.Equal(t, "/Apps", r.URL.Path)

		// Return a list with multiple apps, including the one we want
		response := map[string]interface{}{
			"Items": []interface{}{
				map[string]interface{}{
					"Id":           "test-app-id-2",
					"Name":         "Other App",
					"Description":  "Not the one we want",
					"BusinessUnit": "Sales",
				},
				map[string]interface{}{
					"Id":           "test-app-id-1",
					"Name":         "Test App 1",
					"Description":  "A test application",
					"BusinessUnit": "Engineering",
				},
			},
			"TotalCount": 2,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer ts.Close()

	// Create client pointing to our test server
	c := client.New(ts.URL, "test-key-id", "test-key-secret", 30)

	// Create the tool with readonly disabled (nil means not readonly)
	tool := NewAppGetTool(c, nil)

	// Call the tool
	result, err := tool.Handle(context.Background(), map[string]interface{}{"id": "test-app-id-1"})
	assert.NoError(t, err)
	assert.NotEmpty(t, result)

	// Parse the result
	var output map[string]interface{}
	err = json.Unmarshal([]byte(result), &output)
	assert.NoError(t, err)

	// Check normalized fields
	assert.Equal(t, "test-app-id-1", output["id"])
	assert.Equal(t, "Test App 1", output["name"])
	assert.Equal(t, "A test application", output["description"])
	assert.Equal(t, "Engineering", output["business_unit"])

	// Check raw field is preserved
	raw := output["raw"].(map[string]interface{})
	assert.Equal(t, "test-app-id-1", raw["Id"])
	assert.Equal(t, "Test App 1", raw["Name"])
}

// ---------------------------------------------------------------------------
// AppsSearchTool tests
// ---------------------------------------------------------------------------

func TestAppsSearchTool_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "test-key-id:test-key-secret", r.Header.Get("X-Api-Key"))
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Equal(t, "application/json", r.Header.Get("Accept"))
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/applications/search", r.URL.Path)
		assert.Equal(t, "page=1&pageSize=20", r.URL.RawQuery)

		var body map[string]interface{}
		json.NewDecoder(r.Body).Decode(&body)
		assert.Equal(t, "Test App", body["Name"])

		response := map[string]interface{}{
			"Applications": []interface{}{
				map[string]interface{}{
					"Id":   "app-id-1",
					"Name": "Test App",
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
	tool := NewAppsSearchTool(c, nil)

	result, err := tool.Handle(context.Background(), map[string]interface{}{
		"name": "Test App",
	})
	assert.NoError(t, err)
	assert.NotEmpty(t, result)

	var output map[string]interface{}
	err = json.Unmarshal([]byte(result), &output)
	assert.NoError(t, err)

	apps := output["applications"].([]interface{})
	assert.Len(t, apps, 1)

	first := apps[0].(map[string]interface{})
	assert.Equal(t, "app-id-1", first["id"])
	assert.Equal(t, "Test App", first["name"])

	assert.Equal(t, float64(1), output["page"])
	assert.Equal(t, float64(20), output["page_size"])
	assert.Equal(t, float64(1), output["total_pages"])
	assert.Equal(t, float64(1), output["total_count"])
}

func TestAppGetTool_NotFound(t *testing.T) {
	// Create a mock HTTP server that returns a list without the requested app
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check that the request has the correct headers
		assert.Equal(t, "test-key-id:test-key-secret", r.Header.Get("X-Api-Key"))
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Equal(t, "application/json", r.Header.Get("Accept"))
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/Apps", r.URL.Path) // Uses list endpoint

		// Return a list with different apps (not the one we're looking for)
		response := map[string]interface{}{
			"Items": []interface{}{
				map[string]interface{}{
					"Id":   "other-app-id",
					"Name": "Other App",
				},
			},
			"TotalCount": 1,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer ts.Close()

	// Create client pointing to our test server
	c := client.New(ts.URL, "test-key-id", "test-key-secret", 30)

	// Create the tool with readonly disabled (nil means not readonly)
	tool := NewAppGetTool(c, nil)

	// Call the tool with an ID that's not in the list
	_, err := tool.Handle(context.Background(), map[string]interface{}{"id": "non-existent-id"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "application not found: non-existent-id")
}
