package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/karldane/appscan-asoc-mcp/internal/client"
	"github.com/karldane/appscan-asoc-mcp/internal/model"
	"github.com/karldane/appscan-asoc-mcp/internal/normalize"

	"github.com/karldane/mcp-framework/framework"
	"github.com/mark3labs/mcp-go/mcp"
)

type FindingsListTool struct {
	client *client.Client
	cfg    interface{ ReadOnly() bool }
}

func NewFindingsListTool(c *client.Client, cfg interface{ ReadOnly() bool }) *FindingsListTool {
	return &FindingsListTool{client: c, cfg: cfg}
}

func (t *FindingsListTool) Name() string        { return "findings_list" }
func (t *FindingsListTool) Description() string { return "List findings for an application or scan" }

func (t *FindingsListTool) Schema() mcp.ToolInputSchema {
	return mcp.ToolInputSchema{
		Type: "object",
		Properties: map[string]interface{}{
			"application_id": map[string]interface{}{
				"type":        "string",
				"description": "Application ID",
			},
			"scan_id": map[string]interface{}{
				"type":        "string",
				"description": "Scan ID",
			},
			"page": map[string]interface{}{
				"type":        "integer",
				"description": "Page number",
				"default":     1,
			},
			"page_size": map[string]interface{}{
				"type":        "integer",
				"description": "Results per page",
				"default":     50,
			},
		},
	}
}

func (t *FindingsListTool) Handle(ctx context.Context, args map[string]interface{}) (string, error) {
	page := 1
	pageSize := 50
	if v, ok := args["page"].(float64); ok {
		page = int(v)
	}
	if v, ok := args["page_size"].(float64); ok {
		pageSize = int(v)
	}

	// ASoC v4 uses /Issues/Application/{appId} for listing issues
	// with OData $select and $filter parameters
	applicationID, _ := args["application_id"].(string)
	if applicationID == "" {
		return "", fmt.Errorf("application_id is required")
	}

	path := fmt.Sprintf("/Issues/Application/%s?$skip=%d&$top=%d",
		applicationID, (page-1)*pageSize, pageSize)

	if v, ok := args["scan_id"].(string); ok && v != "" {
		// URL-encode the filter value to handle special characters
		filterValue := fmt.Sprintf("ScanId eq '%s'", v)
		encodedFilter := url.QueryEscape(filterValue)
		path += "&$filter=" + encodedFilter
	}

	resp, err := t.client.Get(path)
	if err != nil {
		return "", fmt.Errorf("list findings: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", client.ParseError(resp)
	}

	var result struct {
		Items []map[string]any `json:"Items"`
		Count int              `json:"TotalCount"`
	}
	if err := client.DecodeJSON(resp, &result); err != nil {
		return "", fmt.Errorf("decode response: %w", err)
	}

	findings := make([]*model.Finding, 0, len(result.Items))
	for _, raw := range result.Items {
		findings = append(findings, normalize.Finding(raw))
	}

	output := map[string]any{
		"findings":    findings,
		"page":        page,
		"page_size":   pageSize,
		"total_count": result.Count,
	}

	b, _ := json.Marshal(output)
	return string(b), nil
}

func (t *FindingsListTool) GetEnforcerProfile() *framework.EnforcerProfile {
	return framework.NewEnforcerProfile(
		framework.WithRisk(framework.RiskMed),
		framework.WithImpact(framework.ImpactRead),
		framework.WithPII(true),
		framework.WithIdempotent(true),
		framework.WithApprovalReq(false),
	)
}

type FindingsSearchTool struct {
	client *client.Client
	cfg    interface{ ReadOnly() bool }
}

func NewFindingsSearchTool(c *client.Client, cfg interface{ ReadOnly() bool }) *FindingsSearchTool {
	return &FindingsSearchTool{client: c, cfg: cfg}
}

func (t *FindingsSearchTool) Name() string { return "findings_search" }
func (t *FindingsSearchTool) Description() string {
	return "Search and filter findings by severity, status, issue type, or text"
}

func (t *FindingsSearchTool) Schema() mcp.ToolInputSchema {
	return mcp.ToolInputSchema{
		Type: "object",
		Properties: map[string]interface{}{
			"application_id": map[string]interface{}{
				"type":        "string",
				"description": "Application ID",
			},
			"severity": map[string]interface{}{
				"type":        "string",
				"description": "Filter by severity (critical, high, medium, low, info)",
			},
			"status": map[string]interface{}{
				"type":        "string",
				"description": "Filter by status (open, fixed, ignored, noise)",
			},
			"issue_type": map[string]interface{}{
				"type":        "string",
				"description": "Filter by issue type",
			},
			"text": map[string]interface{}{
				"type":        "string",
				"description": "Search text in finding title or location",
			},
			"page": map[string]interface{}{
				"type":        "integer",
				"description": "Page number",
				"default":     1,
			},
			"page_size": map[string]interface{}{
				"type":        "integer",
				"description": "Results per page",
				"default":     50,
			},
		},
	}
}

func (t *FindingsSearchTool) Handle(ctx context.Context, args map[string]interface{}) (string, error) {
	page := 1
	pageSize := 50
	if v, ok := args["page"].(float64); ok {
		page = int(v)
	}
	if v, ok := args["page_size"].(float64); ok {
		pageSize = int(v)
	}

	// ASoC v4 uses OData-style filtering on /Issues/Application/{appId}
	applicationID, _ := args["application_id"].(string)
	if applicationID == "" {
		return "", fmt.Errorf("application_id is required")
	}

	// Build OData filter
	var filters []string

	if v, ok := args["severity"].(string); ok && v != "" {
		// Normalize severity for ASoC (capitalize first letter)
		normalizedSeverity := strings.ToUpper(v[:1]) + strings.ToLower(v[1:])
		filters = append(filters, fmt.Sprintf("Severity eq '%s'", normalizedSeverity))
	}
	if v, ok := args["status"].(string); ok && v != "" {
		// ASoC uses "Open", "Fixed", etc.
		normalizedStatus := strings.ToUpper(v[:1]) + strings.ToLower(v[1:])
		filters = append(filters, fmt.Sprintf("Status eq '%s'", normalizedStatus))
	}
	if v, ok := args["issue_type"].(string); ok && v != "" {
		filters = append(filters, fmt.Sprintf("IssueType eq '%s'", v))
	}

	filterClause := ""
	if len(filters) > 0 {
		filterClause = "&$filter=" + url.QueryEscape(strings.Join(filters, " and "))
	}

	path := fmt.Sprintf("/Issues/Application/%s?$skip=%d&$top=%d%s",
		applicationID, (page-1)*pageSize, pageSize, filterClause)

	resp, err := t.client.Get(path)
	if err != nil {
		return "", fmt.Errorf("search findings: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", client.ParseError(resp)
	}

	var result struct {
		Items []map[string]any `json:"Items"`
		Count int              `json:"TotalCount"`
	}
	if err := client.DecodeJSON(resp, &result); err != nil {
		return "", fmt.Errorf("decode response: %w", err)
	}

	findings := make([]*model.Finding, 0, len(result.Items))
	for _, raw := range result.Items {
		findings = append(findings, normalize.Finding(raw))
	}

	output := map[string]any{
		"findings":    findings,
		"page":        page,
		"page_size":   pageSize,
		"total_count": result.Count,
	}

	b, _ := json.Marshal(output)
	return string(b), nil
}

func (t *FindingsSearchTool) GetEnforcerProfile() *framework.EnforcerProfile {
	return framework.NewEnforcerProfile(
		framework.WithRisk(framework.RiskMed),
		framework.WithImpact(framework.ImpactRead),
		framework.WithPII(true),
		framework.WithIdempotent(true),
		framework.WithApprovalReq(false),
	)
}

type FindingGetTool struct {
	client *client.Client
	cfg    interface{ ReadOnly() bool }
}

func NewFindingGetTool(c *client.Client, cfg interface{ ReadOnly() bool }) *FindingGetTool {
	return &FindingGetTool{client: c, cfg: cfg}
}

func (t *FindingGetTool) Name() string        { return "finding_get" }
func (t *FindingGetTool) Description() string { return "Get detailed finding information" }

func (t *FindingGetTool) Schema() mcp.ToolInputSchema {
	return mcp.ToolInputSchema{
		Type: "object",
		Properties: map[string]interface{}{
			"id": map[string]interface{}{
				"type":        "string",
				"description": "Finding ID",
			},
		},
		Required: []string{"id"},
	}
}

func (t *FindingGetTool) Handle(ctx context.Context, args map[string]interface{}) (string, error) {
	id, _ := args["id"].(string)
	if id == "" {
		return "", fmt.Errorf("id is required")
	}

	// ASoC v4 uses /Issues/{id} for getting issue details (JSON)
	path := fmt.Sprintf("/Issues/%s", id)
	resp, err := t.client.Get(path)
	if err != nil {
		return "", fmt.Errorf("get finding: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return "", fmt.Errorf("finding not found: %s", id)
	}
	if resp.StatusCode != 200 {
		return "", client.ParseError(resp)
	}

	var raw map[string]any
	if err := client.DecodeJSON(resp, &raw); err != nil {
		return "", fmt.Errorf("decode response: %w", err)
	}

	finding := normalize.Finding(raw)
	b, _ := json.Marshal(finding)
	return string(b), nil
}

func (t *FindingGetTool) GetEnforcerProfile() *framework.EnforcerProfile {
	return framework.NewEnforcerProfile(
		framework.WithRisk(framework.RiskMed),
		framework.WithImpact(framework.ImpactRead),
		framework.WithPII(true),
		framework.WithIdempotent(true),
		framework.WithApprovalReq(false),
	)
}

type FindingGroupSummaryTool struct {
	client *client.Client
	cfg    interface{ ReadOnly() bool }
}

func NewFindingGroupSummaryTool(c *client.Client, cfg interface{ ReadOnly() bool }) *FindingGroupSummaryTool {
	return &FindingGroupSummaryTool{client: c, cfg: cfg}
}

func (t *FindingGroupSummaryTool) Name() string { return "finding_group_summary" }
func (t *FindingGroupSummaryTool) Description() string {
	return "Aggregate findings by severity, issue type, status, or compliance"
}

func (t *FindingGroupSummaryTool) Schema() mcp.ToolInputSchema {
	return mcp.ToolInputSchema{
		Type: "object",
		Properties: map[string]interface{}{
			"application_id": map[string]interface{}{
				"type":        "string",
				"description": "Application ID",
			},
			"scan_id": map[string]interface{}{
				"type":        "string",
				"description": "Scan ID",
			},
			"group_by": map[string]interface{}{
				"type":        "string",
				"description": "Group by: severity, issue_type, status, compliance",
				"default":     "severity",
			},
		},
		Required: []string{"application_id"},
	}
}

func (t *FindingGroupSummaryTool) Handle(ctx context.Context, args map[string]interface{}) (string, error) {
	appID, _ := args["application_id"].(string)
	scanID, _ := args["scan_id"].(string)
	groupBy, _ := args["group_by"].(string)

	if appID == "" {
		return "", fmt.Errorf("application_id is required")
	}

	if groupBy == "" {
		groupBy = "severity"
	}

	path := fmt.Sprintf("/findings/summary/%s", groupBy)
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
		return "", fmt.Errorf("get finding summary: %w", err)
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

func (t *FindingGroupSummaryTool) GetEnforcerProfile() *framework.EnforcerProfile {
	return framework.NewEnforcerProfile(
		framework.WithRisk(framework.RiskLow),
		framework.WithImpact(framework.ImpactRead),
		framework.WithPII(false),
		framework.WithIdempotent(true),
		framework.WithApprovalReq(false),
	)
}
