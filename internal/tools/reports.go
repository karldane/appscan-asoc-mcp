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

type ReportsListTool struct {
	client *client.Client
	cfg    interface{ ReadOnly() bool }
}

func NewReportsListTool(c *client.Client, cfg interface{ ReadOnly() bool }) *ReportsListTool {
	return &ReportsListTool{client: c, cfg: cfg}
}

func (t *ReportsListTool) Name() string        { return "reports_list" }
func (t *ReportsListTool) Description() string { return "List reports for an application or scan" }

func (t *ReportsListTool) Schema() mcp.ToolInputSchema {
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
				"default":     20,
			},
		},
	}
}

func (t *ReportsListTool) Handle(ctx context.Context, args map[string]interface{}) (string, error) {
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

	path := fmt.Sprintf("/api/v4/reports?page=%d&pageSize=%d", page, pageSize)

	if v, ok := args["application_id"].(string); ok && v != "" {
		path += "&applicationId=" + v
	}
	if v, ok := args["scan_id"].(string); ok && v != "" {
		path += "&scanId=" + v
	}

	resp, err := t.client.Get(path)
	if err != nil {
		return "", fmt.Errorf("list reports: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", client.ParseError(resp)
	}

	var result struct {
		Reports    []map[string]any `json:"Reports"`
		TotalPages int              `json:"TotalPages"`
		TotalCount int              `json:"TotalCount"`
	}
	if err := client.DecodeJSON(resp, &result); err != nil {
		return "", fmt.Errorf("decode response: %w", err)
	}

	reports := make([]*model.Report, 0, len(result.Reports))
	for _, raw := range result.Reports {
		reports = append(reports, normalize.Report(raw))
	}

	output := map[string]any{
		"reports":     reports,
		"page":        page,
		"page_size":   pageSize,
		"total_pages": result.TotalPages,
		"total_count": result.TotalCount,
	}

	b, _ := json.Marshal(output)
	return string(b), nil
}

func (t *ReportsListTool) GetEnforcerProfile() *framework.EnforcerProfile {
	return framework.NewEnforcerProfile(
		framework.WithRisk(framework.RiskMed),
		framework.WithImpact(framework.ImpactRead),
		framework.WithPII(true),
		framework.WithIdempotent(true),
		framework.WithApprovalReq(false),
	)
}

type ReportGenerateTool struct {
	client *client.Client
	cfg    interface{ ReadOnly() bool }
}

func NewReportGenerateTool(c *client.Client, cfg interface{ ReadOnly() bool }) *ReportGenerateTool {
	return &ReportGenerateTool{client: c, cfg: cfg}
}

func (t *ReportGenerateTool) Name() string { return "report_generate" }
func (t *ReportGenerateTool) Description() string {
	return "Request report generation for an application or scan"
}

func (t *ReportGenerateTool) Schema() mcp.ToolInputSchema {
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
			"report_type": map[string]interface{}{
				"type":        "string",
				"description": "Report type (Full, Executive, XML, PDF, HTML)",
				"default":     "PDF",
			},
		},
		Required: []string{"application_id"},
	}
}

func (t *ReportGenerateTool) Handle(ctx context.Context, args map[string]interface{}) (string, error) {
	if readonly.IsReadOnly(t.cfg) {
		return "", fmt.Errorf("server is in readonly mode")
	}

	appID, _ := args["application_id"].(string)
	scanID, _ := args["scan_id"].(string)
	reportType, _ := args["report_type"].(string)

	if appID == "" {
		return "", fmt.Errorf("application_id is required")
	}

	if reportType == "" {
		reportType = "PDF"
	}

	type ReportRequest struct {
		ApplicationID string `json:"ApplicationId"`
		ScanID        string `json:"ScanId,omitempty"`
		ReportType    string `json:"ReportType"`
	}

	req := ReportRequest{
		ApplicationID: appID,
		ReportType:    reportType,
	}
	if scanID != "" {
		req.ScanID = scanID
	}

	resp, err := t.client.Post("/api/v4/reports/generate", req)
	if err != nil {
		return "", fmt.Errorf("generate report: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		return "", client.ParseError(resp)
	}

	var result struct {
		ID     string `json:"Id"`
		Status string `json:"Status"`
	}
	if err := client.DecodeJSON(resp, &result); err != nil {
		return "", fmt.Errorf("decode response: %w", err)
	}

	output := map[string]any{
		"id":      result.ID,
		"status":  result.Status,
		"message": "Report generation started",
	}

	b, _ := json.Marshal(output)
	return string(b), nil
}

func (t *ReportGenerateTool) GetEnforcerProfile() *framework.EnforcerProfile {
	return framework.NewEnforcerProfile(
		framework.WithRisk(framework.RiskHigh),
		framework.WithImpact(framework.ImpactWrite),
		framework.WithPII(false),
		framework.WithIdempotent(false),
		framework.WithApprovalReq(true),
		framework.WithResourceCost(7),
	)
}

type ReportGetTool struct {
	client *client.Client
	cfg    interface{ ReadOnly() bool }
}

func NewReportGetTool(c *client.Client, cfg interface{ ReadOnly() bool }) *ReportGetTool {
	return &ReportGetTool{client: c, cfg: cfg}
}

func (t *ReportGetTool) Name() string        { return "report_get" }
func (t *ReportGetTool) Description() string { return "Get report status and metadata" }

func (t *ReportGetTool) Schema() mcp.ToolInputSchema {
	return mcp.ToolInputSchema{
		Type: "object",
		Properties: map[string]interface{}{
			"id": map[string]interface{}{
				"type":        "string",
				"description": "Report ID",
			},
		},
		Required: []string{"id"},
	}
}

func (t *ReportGetTool) Handle(ctx context.Context, args map[string]interface{}) (string, error) {
	if readonly.IsReadOnly(t.cfg) {
		return "", fmt.Errorf("server is in readonly mode")
	}

	id, _ := args["id"].(string)
	if id == "" {
		return "", fmt.Errorf("id is required")
	}

	path := fmt.Sprintf("/api/v4/reports/%s", id)
	resp, err := t.client.Get(path)
	if err != nil {
		return "", fmt.Errorf("get report: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return "", fmt.Errorf("report not found: %s", id)
	}
	if resp.StatusCode != 200 {
		return "", client.ParseError(resp)
	}

	var raw map[string]any
	if err := client.DecodeJSON(resp, &raw); err != nil {
		return "", fmt.Errorf("decode response: %w", err)
	}

	report := normalize.Report(raw)
	b, _ := json.Marshal(report)
	return string(b), nil
}

func (t *ReportGetTool) GetEnforcerProfile() *framework.EnforcerProfile {
	return framework.NewEnforcerProfile(
		framework.WithRisk(framework.RiskMed),
		framework.WithImpact(framework.ImpactRead),
		framework.WithPII(true),
		framework.WithIdempotent(true),
		framework.WithApprovalReq(false),
	)
}
