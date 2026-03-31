package tools

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"

	"appscan-asoc-mcp/internal/client"
	"appscan-asoc-mcp/internal/readonly"

	"github.com/karldane/mcp-framework/framework"
	"github.com/mark3labs/mcp-go/mcp"
)

type FilesUploadTool struct {
	client *client.Client
	cfg    interface{ ReadOnly() bool }
}

func NewFilesUploadTool(c *client.Client, cfg interface{ ReadOnly() bool }) *FilesUploadTool {
	return &FilesUploadTool{client: c, cfg: cfg}
}

func (t *FilesUploadTool) Name() string        { return "files_upload" }
func (t *FilesUploadTool) Description() string { return "Upload a scan, template, or config file" }

func (t *FilesUploadTool) Schema() mcp.ToolInputSchema {
	return mcp.ToolInputSchema{
		Type: "object",
		Properties: map[string]interface{}{
			"file_name": map[string]interface{}{
				"type":        "string",
				"description": "Name of the file",
			},
			"file_content": map[string]interface{}{
				"type":        "string",
				"description": "Base64-encoded file content",
			},
		},
		Required: []string{"file_name", "file_content"},
	}
}

func (t *FilesUploadTool) Handle(ctx context.Context, args map[string]interface{}) (string, error) {
	if readonly.IsReadOnly(t.cfg) {
		return "", fmt.Errorf("server is in readonly mode")
	}

	fileName, _ := args["file_name"].(string)
	fileContent, _ := args["file_content"].(string)

	if fileName == "" || fileContent == "" {
		return "", fmt.Errorf("file_name and file_content are required")
	}

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	part, err := writer.CreateFormFile("file", fileName)
	if err != nil {
		return "", fmt.Errorf("create form file: %w", err)
	}

	decoded, err := base64.StdEncoding.DecodeString(fileContent)
	if err != nil {
		return "", fmt.Errorf("decode base64: %w", err)
	}
	if _, err := part.Write(decoded); err != nil {
		return "", fmt.Errorf("write file content: %w", err)
	}

	writer.Close()

	req, err := http.NewRequest("POST", t.client.BaseURL()+"/api/v4/files/upload", &body)
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Authorization", t.client.AuthHeader())
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := t.client.DoCustom("POST", "/api/v4/files/upload",
		map[string]string{"Content-Type": writer.FormDataContentType()},
		&body)
	if err != nil {
		return "", fmt.Errorf("upload file: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		return "", client.ParseError(resp)
	}

	var result struct {
		ID   string `json:"Id"`
		Name string `json:"Name"`
		Size int64  `json:"Size"`
	}
	if err := client.DecodeJSON(resp, &result); err != nil {
		return "", fmt.Errorf("decode response: %w", err)
	}

	output := map[string]any{
		"id":       result.ID,
		"name":     result.Name,
		"size":     result.Size,
		"file_url": fmt.Sprintf("%s/api/v4/files/%s", t.client.BaseURL(), result.ID),
	}

	b, _ := json.Marshal(output)
	return string(b), nil
}

func (t *FilesUploadTool) GetEnforcerProfile() *framework.EnforcerProfile {
	return framework.NewEnforcerProfile(
		framework.WithRisk(framework.RiskHigh),
		framework.WithImpact(framework.ImpactWrite),
		framework.WithPII(true),
		framework.WithIdempotent(false),
		framework.WithApprovalReq(true),
	)
}

type FileGetTool struct {
	client *client.Client
	cfg    interface{ ReadOnly() bool }
}

func NewFileGetTool(c *client.Client, cfg interface{ ReadOnly() bool }) *FileGetTool {
	return &FileGetTool{client: c, cfg: cfg}
}

func (t *FileGetTool) Name() string        { return "file_get" }
func (t *FileGetTool) Description() string { return "Get file metadata by ID" }

func (t *FileGetTool) Schema() mcp.ToolInputSchema {
	return mcp.ToolInputSchema{
		Type: "object",
		Properties: map[string]interface{}{
			"id": map[string]interface{}{
				"type":        "string",
				"description": "File ID",
			},
		},
		Required: []string{"id"},
	}
}

func (t *FileGetTool) Handle(ctx context.Context, args map[string]interface{}) (string, error) {
	if readonly.IsReadOnly(t.cfg) {
		return "", fmt.Errorf("server is in readonly mode")
	}

	id, _ := args["id"].(string)
	if id == "" {
		return "", fmt.Errorf("id is required")
	}

	path := fmt.Sprintf("/api/v4/files/%s", id)
	resp, err := t.client.Get(path)
	if err != nil {
		return "", fmt.Errorf("get file: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return "", fmt.Errorf("file not found: %s", id)
	}
	if resp.StatusCode != 200 {
		return "", client.ParseError(resp)
	}

	var raw map[string]any
	if err := client.DecodeJSON(resp, &raw); err != nil {
		return "", fmt.Errorf("decode response: %w", err)
	}

	b, _ := json.Marshal(raw)
	return string(b), nil
}

func (t *FileGetTool) GetEnforcerProfile() *framework.EnforcerProfile {
	return framework.NewEnforcerProfile(
		framework.WithRisk(framework.RiskMed),
		framework.WithImpact(framework.ImpactRead),
		framework.WithPII(false),
		framework.WithIdempotent(true),
		framework.WithApprovalReq(false),
	)
}
