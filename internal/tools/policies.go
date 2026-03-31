package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"appscan-asoc-mcp/internal/client"
	"appscan-asoc-mcp/internal/readonly"

	"github.com/karldane/mcp-framework/framework"
	"github.com/mark3labs/mcp-go/mcp"
)

type AssetGroupsListTool struct {
	client *client.Client
	cfg    interface{ ReadOnly() bool }
}

func NewAssetGroupsListTool(c *client.Client, cfg interface{ ReadOnly() bool }) *AssetGroupsListTool {
	return &AssetGroupsListTool{client: c, cfg: cfg}
}

func (t *AssetGroupsListTool) Name() string        { return "asset_groups_list" }
func (t *AssetGroupsListTool) Description() string { return "List asset groups" }

func (t *AssetGroupsListTool) Schema() mcp.ToolInputSchema {
	return mcp.ToolInputSchema{
		Type: "object",
		Properties: map[string]interface{}{
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

func (t *AssetGroupsListTool) Handle(ctx context.Context, args map[string]interface{}) (string, error) {
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

	path := fmt.Sprintf("/api/v4/assetgroups?page=%d&pageSize=%d", page, pageSize)
	resp, err := t.client.Get(path)
	if err != nil {
		return "", fmt.Errorf("list asset groups: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", client.ParseError(resp)
	}

	var result struct {
		AssetGroups []map[string]any `json:"AssetGroups"`
		TotalPages  int              `json:"TotalPages"`
		TotalCount  int              `json:"TotalCount"`
	}
	if err := client.DecodeJSON(resp, &result); err != nil {
		return "", fmt.Errorf("decode response: %w", err)
	}

	output := map[string]any{
		"asset_groups": result.AssetGroups,
		"page":         page,
		"page_size":    pageSize,
		"total_pages":  result.TotalPages,
		"total_count":  result.TotalCount,
	}

	b, _ := json.Marshal(output)
	return string(b), nil
}

func (t *AssetGroupsListTool) GetEnforcerProfile() *framework.EnforcerProfile {
	return framework.NewEnforcerProfile(
		framework.WithRisk(framework.RiskLow),
		framework.WithImpact(framework.ImpactRead),
		framework.WithPII(false),
		framework.WithIdempotent(true),
		framework.WithApprovalReq(false),
	)
}

type PoliciesListTool struct {
	client *client.Client
	cfg    interface{ ReadOnly() bool }
}

func NewPoliciesListTool(c *client.Client, cfg interface{ ReadOnly() bool }) *PoliciesListTool {
	return &PoliciesListTool{client: c, cfg: cfg}
}

func (t *PoliciesListTool) Name() string        { return "policies_list" }
func (t *PoliciesListTool) Description() string { return "List security policies" }

func (t *PoliciesListTool) Schema() mcp.ToolInputSchema {
	return mcp.ToolInputSchema{
		Type: "object",
		Properties: map[string]interface{}{
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

func (t *PoliciesListTool) Handle(ctx context.Context, args map[string]interface{}) (string, error) {
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

	path := fmt.Sprintf("/api/v4/policies?page=%d&pageSize=%d", page, pageSize)
	resp, err := t.client.Get(path)
	if err != nil {
		return "", fmt.Errorf("list policies: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", client.ParseError(resp)
	}

	var result struct {
		Policies   []map[string]any `json:"Policies"`
		TotalPages int              `json:"TotalPages"`
		TotalCount int              `json:"TotalCount"`
	}
	if err := client.DecodeJSON(resp, &result); err != nil {
		return "", fmt.Errorf("decode response: %w", err)
	}

	output := map[string]any{
		"policies":    result.Policies,
		"page":        page,
		"page_size":   pageSize,
		"total_pages": result.TotalPages,
		"total_count": result.TotalCount,
	}

	b, _ := json.Marshal(output)
	return string(b), nil
}

func (t *PoliciesListTool) GetEnforcerProfile() *framework.EnforcerProfile {
	return framework.NewEnforcerProfile(
		framework.WithRisk(framework.RiskLow),
		framework.WithImpact(framework.ImpactRead),
		framework.WithPII(false),
		framework.WithIdempotent(true),
		framework.WithApprovalReq(false),
	)
}

type ComplianceSummaryTool struct {
	client *client.Client
	cfg    interface{ ReadOnly() bool }
}

func NewComplianceSummaryTool(c *client.Client, cfg interface{ ReadOnly() bool }) *ComplianceSummaryTool {
	return &ComplianceSummaryTool{client: c, cfg: cfg}
}

func (t *ComplianceSummaryTool) Name() string { return "compliance_summary" }
func (t *ComplianceSummaryTool) Description() string {
	return "Get compliance/policy summary for an application or scan"
}

func (t *ComplianceSummaryTool) Schema() mcp.ToolInputSchema {
	return mcp.ToolInputSchema{
		Type: "object",
		Properties: map[string]interface{}{
			"application_id": map[string]interface{}{
				"type":        "string",
				"description": "Application ID",
			},
			"scan_id": map[string]interface{}{
				"type":        "string",
				"description": "Scan ID (optional)",
			},
		},
		Required: []string{"application_id"},
	}
}

func (t *ComplianceSummaryTool) Handle(ctx context.Context, args map[string]interface{}) (string, error) {
	if readonly.IsReadOnly(t.cfg) {
		return "", fmt.Errorf("server is in readonly mode")
	}

	appID, _ := args["application_id"].(string)
	scanID, _ := args["scan_id"].(string)

	if appID == "" {
		return "", fmt.Errorf("application_id is required")
	}

	path := fmt.Sprintf("/api/v4/compliance/summary")
	type SummaryRequest struct {
		ApplicationID string `json:"ApplicationId"`
		ScanID        string `json:"ScanId,omitempty"`
	}
	req := SummaryRequest{ApplicationID: appID}
	if scanID != "" {
		req.ScanID = scanID
	}

	resp, err := t.client.Post(path, req)
	if err != nil {
		return "", fmt.Errorf("get compliance summary: %w", err)
	}
	defer resp.Body.Close()

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

func (t *ComplianceSummaryTool) GetEnforcerProfile() *framework.EnforcerProfile {
	return framework.NewEnforcerProfile(
		framework.WithRisk(framework.RiskLow),
		framework.WithImpact(framework.ImpactRead),
		framework.WithPII(false),
		framework.WithIdempotent(true),
		framework.WithApprovalReq(false),
	)
}
