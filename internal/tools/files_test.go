package tools

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"mime"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/karldane/appscan-asoc-mcp/internal/client"
	"github.com/stretchr/testify/assert"
)

// ---------------------------------------------------------------------------
// FilesUploadTool tests
// ---------------------------------------------------------------------------

func TestFilesUploadTool_Success(t *testing.T) {
	fileContent := []byte("fake scan file content")
	fileContentB64 := base64.StdEncoding.EncodeToString(fileContent)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "test-key-id:test-key-secret", r.Header.Get("X-Api-Key"))
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/files/upload", r.URL.Path)

		// Verify multipart content-type
		ct := r.Header.Get("Content-Type")
		mediaType, params, err := mime.ParseMediaType(ct)
		assert.NoError(t, err)
		assert.Equal(t, "multipart/form-data", mediaType)
		assert.NotEmpty(t, params["boundary"])

		// Parse multipart body and check file field
		mr := multipart.NewReader(r.Body, params["boundary"])
		part, err := mr.NextPart()
		assert.NoError(t, err)
		assert.Equal(t, "file", part.FormName())
		assert.Equal(t, "test.scan", part.FileName())

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"Id":   "file-id-1",
			"Name": "test.scan",
			"Size": 1024,
		})
	}))
	defer ts.Close()

	c := client.New(ts.URL, "test-key-id", "test-key-secret", 30)
	tool := NewFilesUploadTool(c, nil)

	result, err := tool.Handle(context.Background(), map[string]interface{}{
		"file_name":    "test.scan",
		"file_content": fileContentB64,
	})
	assert.NoError(t, err)
	assert.NotEmpty(t, result)

	var output map[string]interface{}
	err = json.Unmarshal([]byte(result), &output)
	assert.NoError(t, err)

	assert.Equal(t, "file-id-1", output["id"])
	assert.Equal(t, "test.scan", output["name"])
	assert.Equal(t, float64(1024), output["size"])
	assert.Contains(t, output["file_url"].(string), "file-id-1")
}

func TestFilesUploadTool_ReadOnly(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("request should not be made in readonly mode")
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	c := client.New(ts.URL, "test-key-id", "test-key-secret", 30)
	tool := NewFilesUploadTool(c, &readonlyFlag{true})

	_, err := tool.Handle(context.Background(), map[string]interface{}{
		"file_name":    "test.scan",
		"file_content": base64.StdEncoding.EncodeToString([]byte("data")),
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "readonly mode")
}

func TestFilesUploadTool_MissingFileName(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("request should not be made when args are missing")
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	c := client.New(ts.URL, "test-key-id", "test-key-secret", 30)
	tool := NewFilesUploadTool(c, nil)

	_, err := tool.Handle(context.Background(), map[string]interface{}{
		"file_content": base64.StdEncoding.EncodeToString([]byte("data")),
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "file_name and file_content are required")
}

func TestFilesUploadTool_MissingFileContent(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("request should not be made when args are missing")
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	c := client.New(ts.URL, "test-key-id", "test-key-secret", 30)
	tool := NewFilesUploadTool(c, nil)

	_, err := tool.Handle(context.Background(), map[string]interface{}{
		"file_name": "test.scan",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "file_name and file_content are required")
}

func TestFilesUploadTool_BadBase64(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("request should not be made with invalid base64")
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	c := client.New(ts.URL, "test-key-id", "test-key-secret", 30)
	tool := NewFilesUploadTool(c, nil)

	_, err := tool.Handle(context.Background(), map[string]interface{}{
		"file_name":    "test.scan",
		"file_content": "!!!not-valid-base64!!!",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "decode base64")
}

// ---------------------------------------------------------------------------
// FileGetTool tests
// ---------------------------------------------------------------------------

func TestFileGetTool_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "test-key-id:test-key-secret", r.Header.Get("X-Api-Key"))
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Equal(t, "application/json", r.Header.Get("Accept"))
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/files/file-id-1", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"Id":   "file-id-1",
			"Name": "template.scant",
			"Size": 512,
		})
	}))
	defer ts.Close()

	c := client.New(ts.URL, "test-key-id", "test-key-secret", 30)
	tool := NewFileGetTool(c, nil)

	result, err := tool.Handle(context.Background(), map[string]interface{}{"id": "file-id-1"})
	assert.NoError(t, err)
	assert.NotEmpty(t, result)

	var output map[string]interface{}
	err = json.Unmarshal([]byte(result), &output)
	assert.NoError(t, err)

	// Response is raw JSON passthrough
	assert.Equal(t, "file-id-1", output["Id"])
	assert.Equal(t, "template.scant", output["Name"])
	assert.Equal(t, float64(512), output["Size"])
}

func TestFileGetTool_NotFound(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/files/nonexistent-id", r.URL.Path)
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	c := client.New(ts.URL, "test-key-id", "test-key-secret", 30)
	tool := NewFileGetTool(c, nil)

	_, err := tool.Handle(context.Background(), map[string]interface{}{"id": "nonexistent-id"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "file not found: nonexistent-id")
}

func TestFileGetTool_MissingID(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("request should not be made when id is missing")
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	c := client.New(ts.URL, "test-key-id", "test-key-secret", 30)
	tool := NewFileGetTool(c, nil)

	_, err := tool.Handle(context.Background(), map[string]interface{}{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "id is required")
}

func TestFileGetTool_ReadOnly(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("request should not be made in readonly mode")
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	c := client.New(ts.URL, "test-key-id", "test-key-secret", 30)
	tool := NewFileGetTool(c, &readonlyFlag{true})

	_, err := tool.Handle(context.Background(), map[string]interface{}{"id": "file-id-1"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "readonly mode")
}
