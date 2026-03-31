package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"appscan-asoc-mcp/internal/client"
	"appscan-asoc-mcp/internal/model"
	"appscan-asoc-mcp/internal/normalize"
	"appscan-asoc-mcp/internal/readonly"

	"github.com/karldane/mcp-framework/framework"
	"github.com/mark3labs/mcp-go/mcp"
)

type AppsListTool struct {
	client *client.Client
	cfg    interface{ ReadOnly() bool }
}

func NewAppsListTool(c *client.Client, cfg interface{ ReadOnly() bool }) *AppsListTool {
	return &AppsListTool{client: c, cfg: cfg}
}

func (t *AppsListTool) Name() string { return "apps_list" }
func (t *AppsListTool) Description() string {
	return "List applications with optional filtering and pagination"
}

func (t *AppsListTool) Schema() mcp.ToolInputSchema {
	return mcp.ToolInputSchema{
		Type: "object",
		Properties: map[string]interface{}{
			"page": map[string]interface{}{
				"type":        "integer",
				"description": "Page number (1-indexed)",
				"default":     1,
			},
			"page_size": map[string]interface{}{
				"type":        "integer",
				"description": "Number of results per page",
				"default":     20,
			},
		},
	}
}

func (t *AppsListTool) Handle(ctx context.Context, args map[string]interface{}) (string, error) {
	if readonly.IsReadOnly(t.cfg) {
		return "", fmt.Errorf("server is in readonly mode")
	}

	page := 1
	pageSize := 20
	if v, ok := args["page"].(float64); ok {
		page = int(v)
	}
	if v, ok := args["page_size"].(float64); ok {
		pageSize = int(v)
	}

	path := fmt.Sprintf("/api/v4/applications?page=%d&pageSize=%d", page, pageSize)
	resp, err := t.client.Get(path)
	if err != nil {
		return "", fmt.Errorf("list applications: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", client.ParseError(resp)
	}

	var result struct {
		Applications []map[string]any `json:"Applications"`
		TotalPages   int              `json:"TotalPages"`
		TotalCount   int              `json:"TotalCount"`
	}
	if err := client.DecodeJSON(resp, &result); err != nil {
		return "", fmt.Errorf("decode response: %w", err)
	}

	apps := make([]*model.Application, 0, len(result.Applications))
	for _, raw := range result.Applications {
		apps = append(apps, normalize.Application(raw))
	}

	output := map[string]any{
		"applications": apps,
		"page":         page,
		"page_size":    pageSize,
		"total_pages":  result.TotalPages,
		"total_count":  result.TotalCount,
	}

	b, _ := json.Marshal(output)
	return string(b), nil
}

func (t *AppsListTool) GetEnforcerProfile() *framework.EnforcerProfile {
	return framework.NewEnforcerProfile(
		framework.WithRisk(framework.RiskMed),
		framework.WithImpact(framework.ImpactRead),
		framework.WithPII(true),
		framework.WithIdempotent(true),
		framework.WithApprovalReq(false),
	)
}

type AppGetTool struct {
	client *client.Client
	cfg    interface{ ReadOnly() bool }
}

func NewAppGetTool(c *client.Client, cfg interface{ ReadOnly() bool }) *AppGetTool {
	return &AppGetTool{client: c, cfg: cfg}
}

func (t *AppGetTool) Name() string        { return "app_get" }
func (t *AppGetTool) Description() string { return "Get detailed application metadata by ID" }

func (t *AppGetTool) Schema() mcp.ToolInputSchema {
	return mcp.ToolInputSchema{
		Type: "object",
		Properties: map[string]interface{}{
			"id": map[string]interface{}{
				"type":        "string",
				"description": "Application ID",
			},
		},
		Required: []string{"id"},
	}
}

func (t *AppGetTool) Handle(ctx context.Context, args map[string]interface{}) (string, error) {
	if readonly.IsReadOnly(t.cfg) {
		return "", fmt.Errorf("server is in readonly mode")
	}

	id, _ := args["id"].(string)
	if id == "" {
		return "", fmt.Errorf("id is required")
	}

	path := fmt.Sprintf("/api/v4/applications/%s", id)
	resp, err := t.client.Get(path)
	if err != nil {
		return "", fmt.Errorf("get application: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return "", fmt.Errorf("application not found: %s", id)
	}
	if resp.StatusCode != 200 {
		return "", client.ParseError(resp)
	}

	var raw map[string]any
	if err := client.DecodeJSON(resp, &raw); err != nil {
		return "", fmt.Errorf("decode response: %w", err)
	}

	app := normalize.Application(raw)
	b, _ := json.Marshal(app)
	return string(b), nil
}

func (t *AppGetTool) GetEnforcerProfile() *framework.EnforcerProfile {
	return framework.NewEnforcerProfile(
		framework.WithRisk(framework.RiskMed),
		framework.WithImpact(framework.ImpactRead),
		framework.WithPII(true),
		framework.WithIdempotent(true),
		framework.WithApprovalReq(false),
	)
}

type AppsSearchTool struct {
	client *client.Client
	cfg    interface{ ReadOnly() bool }
}

func NewAppsSearchTool(c *client.Client, cfg interface{ ReadOnly() bool }) *AppsSearchTool {
	return &AppsSearchTool{client: c, cfg: cfg}
}

func (t *AppsSearchTool) Name() string { return "apps_search" }
func (t *AppsSearchTool) Description() string {
	return "Search applications by name, tag, business unit, or status"
}

func (t *AppsSearchTool) Schema() mcp.ToolInputSchema {
	return mcp.ToolInputSchema{
		Type: "object",
		Properties: map[string]interface{}{
			"name": map[string]interface{}{
				"type":        "string",
				"description": "Search by application name (partial match)",
			},
			"tag": map[string]interface{}{
				"type":        "string",
				"description": "Filter by tag",
			},
			"business_unit": map[string]interface{}{
				"type":        "string",
				"description": "Filter by business unit",
			},
			"page": map[string]interface{}{
				"type":        "integer",
				"description": "Page number",
				"default":     1,
			},
			"page_size": map[string]interface{}{
				"type":        "integer",
				"description": "Results per page",
				"default":     20,
			},
		},
	}
}

func (t *AppsSearchTool) Handle(ctx context.Context, args map[string]interface{}) (string, error) {
	if readonly.IsReadOnly(t.cfg) {
		return "", fmt.Errorf("server is in readonly mode")
	}

	page := 1
	pageSize := 20
	if v, ok := args["page"].(float64); ok {
		page = int(v)
	}
	if v, ok := args["page_size"].(float64); ok {
		pageSize = int(v)
	}

	path := fmt.Sprintf("/api/v4/applications/search?page=%d&pageSize=%d", page, pageSize)

	type SearchRequest struct {
		Name         string `json:"Name,omitempty"`
		Tag          string `json:"Tag,omitempty"`
		BusinessUnit string `json:"BusinessUnit,omitempty"`
	}

	req := SearchRequest{}
	if v, ok := args["name"].(string); ok {
		req.Name = v
	}
	if v, ok := args["tag"].(string); ok {
		req.Tag = v
	}
	if v, ok := args["business_unit"].(string); ok {
		req.BusinessUnit = v
	}

	resp, err := t.client.Post(path, req)
	if err != nil {
		return "", fmt.Errorf("search applications: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", client.ParseError(resp)
	}

	var result struct {
		Applications []map[string]any `json:"Applications"`
		TotalPages   int              `json:"TotalPages"`
		TotalCount   int              `json:"TotalCount"`
	}
	if err := client.DecodeJSON(resp, &result); err != nil {
		return "", fmt.Errorf("decode response: %w", err)
	}

	apps := make([]*model.Application, 0, len(result.Applications))
	for _, raw := range result.Applications {
		apps = append(apps, normalize.Application(raw))
	}

	output := map[string]any{
		"applications": apps,
		"page":         page,
		"page_size":    pageSize,
		"total_pages":  result.TotalPages,
		"total_count":  result.TotalCount,
	}

	b, _ := json.Marshal(output)
	return string(b), nil
}

func (t *AppsSearchTool) GetEnforcerProfile() *framework.EnforcerProfile {
	return framework.NewEnforcerProfile(
		framework.WithRisk(framework.RiskMed),
		framework.WithImpact(framework.ImpactRead),
		framework.WithPII(true),
		framework.WithIdempotent(true),
		framework.WithApprovalReq(false),
	)
}
