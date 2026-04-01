package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/karldane/appscan-asoc-mcp/internal/client"
	"github.com/karldane/appscan-asoc-mcp/internal/model"
	"github.com/karldane/appscan-asoc-mcp/internal/normalize"
	"github.com/karldane/appscan-asoc-mcp/internal/readonly"

	"github.com/karldane/mcp-framework/framework"
	"github.com/mark3labs/mcp-go/mcp"
)

type ScansListTool struct {
	client *client.Client
	cfg    interface{ ReadOnly() bool }
}

func NewScansListTool(c *client.Client, cfg interface{ ReadOnly() bool }) *ScansListTool {
	return &ScansListTool{client: c, cfg: cfg}
}

func (t *ScansListTool) Name() string { return "scans_list" }
func (t *ScansListTool) Description() string {
	return "List scans with filters by app, family, state, and date"
}

func (t *ScansListTool) Schema() mcp.ToolInputSchema {
	return mcp.ToolInputSchema{
		Type: "object",
		Properties: map[string]interface{}{
			"application_id": map[string]interface{}{
				"type":        "string",
				"description": "Filter by application ID",
			},
			"scan_family": map[string]interface{}{
				"type":        "string",
				"description": "Filter by scan family (DAST, SAST, etc.)",
			},
			"status": map[string]interface{}{
				"type":        "string",
				"description": "Filter by status (queued, running, completed, failed)",
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

func (t *ScansListTool) Handle(ctx context.Context, args map[string]interface{}) (string, error) {
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

	path := fmt.Sprintf("/scans?page=%d&pageSize=%d", page, pageSize)

	if v, ok := args["application_id"].(string); ok && v != "" {
		path += "&applicationId=" + v
	}
	if v, ok := args["scan_family"].(string); ok && v != "" {
		path += "&scanType=" + v
	}
	if v, ok := args["status"].(string); ok && v != "" {
		path += "&state=" + v
	}

	resp, err := t.client.Get(path)
	if err != nil {
		return "", fmt.Errorf("list scans: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", client.ParseError(resp)
	}

	var result struct {
		Scans      []map[string]any `json:"Scans"`
		TotalPages int              `json:"TotalPages"`
		TotalCount int              `json:"TotalCount"`
	}
	if err := client.DecodeJSON(resp, &result); err != nil {
		return "", fmt.Errorf("decode response: %w", err)
	}

	scans := make([]*model.Scan, 0, len(result.Scans))
	for _, raw := range result.Scans {
		scans = append(scans, normalize.Scan(raw))
	}

	output := map[string]any{
		"scans":       scans,
		"page":        page,
		"page_size":   pageSize,
		"total_pages": result.TotalPages,
		"total_count": result.TotalCount,
	}

	b, _ := json.Marshal(output)
	return string(b), nil
}

func (t *ScansListTool) GetEnforcerProfile() framework.EnforcerProfile {
	return framework.NewEnforcerProfile(
		framework.WithRisk(framework.RiskMed),
		framework.WithImpact(framework.ImpactRead),
		framework.WithPII(true),
		framework.WithIdempotent(true),
		framework.WithApprovalReq(false),
	)
}

type ScanGetTool struct {
	client *client.Client
	cfg    interface{ ReadOnly() bool }
}

func NewScanGetTool(c *client.Client, cfg interface{ ReadOnly() bool }) *ScanGetTool {
	return &ScanGetTool{client: c, cfg: cfg}
}

func (t *ScanGetTool) Name() string        { return "scan_get" }
func (t *ScanGetTool) Description() string { return "Get detailed scan metadata" }

func (t *ScanGetTool) Schema() mcp.ToolInputSchema {
	return mcp.ToolInputSchema{
		Type: "object",
		Properties: map[string]interface{}{
			"id": map[string]interface{}{
				"type":        "string",
				"description": "Scan ID",
			},
		},
		Required: []string{"id"},
	}
}

func (t *ScanGetTool) Handle(ctx context.Context, args map[string]interface{}) (string, error) {
	if readonly.IsReadOnly(t.cfg) {
		return "", fmt.Errorf("server is in readonly mode")
	}

	id, _ := args["id"].(string)
	if id == "" {
		return "", fmt.Errorf("id is required")
	}

	path := fmt.Sprintf("/scans/%s", id)
	resp, err := t.client.Get(path)
	if err != nil {
		return "", fmt.Errorf("get scan: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return "", fmt.Errorf("scan not found: %s", id)
	}
	if resp.StatusCode != 200 {
		return "", client.ParseError(resp)
	}

	var raw map[string]any
	if err := client.DecodeJSON(resp, &raw); err != nil {
		return "", fmt.Errorf("decode response: %w", err)
	}

	scan := normalize.Scan(raw)
	b, _ := json.Marshal(scan)
	return string(b), nil
}

func (t *ScanGetTool) GetEnforcerProfile() framework.EnforcerProfile {
	return framework.NewEnforcerProfile(
		framework.WithRisk(framework.RiskMed),
		framework.WithImpact(framework.ImpactRead),
		framework.WithPII(true),
		framework.WithIdempotent(true),
		framework.WithApprovalReq(false),
	)
}

type ScanStatusTool struct {
	client *client.Client
	cfg    interface{ ReadOnly() bool }
}

func NewScanStatusTool(c *client.Client, cfg interface{ ReadOnly() bool }) *ScanStatusTool {
	return &ScanStatusTool{client: c, cfg: cfg}
}

func (t *ScanStatusTool) Name() string { return "scan_status" }
func (t *ScanStatusTool) Description() string {
	return "Return normalized scan state and queue details"
}

func (t *ScanStatusTool) Schema() mcp.ToolInputSchema {
	return mcp.ToolInputSchema{
		Type: "object",
		Properties: map[string]interface{}{
			"id": map[string]interface{}{
				"type":        "string",
				"description": "Scan ID",
			},
		},
		Required: []string{"id"},
	}
}

func (t *ScanStatusTool) Handle(ctx context.Context, args map[string]interface{}) (string, error) {
	if readonly.IsReadOnly(t.cfg) {
		return "", fmt.Errorf("server is in readonly mode")
	}

	id, _ := args["id"].(string)
	if id == "" {
		return "", fmt.Errorf("id is required")
	}

	path := fmt.Sprintf("/scans/%s", id)
	resp, err := t.client.Get(path)
	if err != nil {
		return "", fmt.Errorf("get scan status: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return "", fmt.Errorf("scan not found: %s", id)
	}
	if resp.StatusCode != 200 {
		return "", client.ParseError(resp)
	}

	var raw map[string]any
	if err := client.DecodeJSON(resp, &raw); err != nil {
		return "", fmt.Errorf("decode response: %w", err)
	}

	scan := normalize.Scan(raw)

	output := map[string]any{
		"id":           scan.ID,
		"status":       scan.Status,
		"queue_state":  scan.QueueState,
		"submitted_at": scan.SubmittedAt,
		"started_at":   scan.StartedAt,
		"completed_at": scan.CompletedAt,
	}

	b, _ := json.Marshal(output)
	return string(b), nil
}

func (t *ScanStatusTool) GetEnforcerProfile() framework.EnforcerProfile {
	return framework.NewEnforcerProfile(
		framework.WithRisk(framework.RiskLow),
		framework.WithImpact(framework.ImpactRead),
		framework.WithPII(false),
		framework.WithIdempotent(true),
		framework.WithApprovalReq(false),
	)
}

type DASTScanStartTool struct {
	client *client.Client
	cfg    interface{ ReadOnly() bool }
}

func NewDASTScanStartTool(c *client.Client, cfg interface{ ReadOnly() bool }) *DASTScanStartTool {
	return &DASTScanStartTool{client: c, cfg: cfg}
}

func (t *DASTScanStartTool) Name() string { return "dast_scan_start" }
func (t *DASTScanStartTool) Description() string {
	return "Start a DAST scan for an application or target URL"
}

func (t *DASTScanStartTool) Schema() mcp.ToolInputSchema {
	return mcp.ToolInputSchema{
		Type: "object",
		Properties: map[string]interface{}{
			"application_id": map[string]interface{}{
				"type":        "string",
				"description": "Application ID to scan",
			},
			"target_url": map[string]interface{}{
				"type":        "string",
				"description": "Target URL to scan",
			},
			"scan_name": map[string]interface{}{
				"type":        "string",
				"description": "Optional name for the scan",
			},
			"policy_id": map[string]interface{}{
				"type":        "string",
				"description": "Optional policy ID to use",
			},
		},
		Required: []string{"application_id", "target_url"},
	}
}

func (t *DASTScanStartTool) Handle(ctx context.Context, args map[string]interface{}) (string, error) {
	if readonly.IsReadOnly(t.cfg) {
		return "", fmt.Errorf("server is in readonly mode")
	}

	appID, _ := args["application_id"].(string)
	targetURL, _ := args["target_url"].(string)

	if appID == "" || targetURL == "" {
		return "", fmt.Errorf("application_id and target_url are required")
	}

	scanName, _ := args["scan_name"].(string)
	policyID, _ := args["policy_id"].(string)

	type ScanRequest struct {
		ApplicationID string `json:"ApplicationId"`
		URL           string `json:"Url"`
		Name          string `json:"Name,omitempty"`
		PolicyID      string `json:"PolicyId,omitempty"`
	}

	req := ScanRequest{
		ApplicationID: appID,
		URL:           targetURL,
	}
	if scanName != "" {
		req.Name = scanName
	}
	if policyID != "" {
		req.PolicyID = policyID
	}

	resp, err := t.client.Post("/scans/dast", req)
	if err != nil {
		return "", fmt.Errorf("start scan: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		return "", client.ParseError(resp)
	}

	var raw map[string]any
	if err := client.DecodeJSON(resp, &raw); err != nil {
		return "", fmt.Errorf("decode response: %w", err)
	}

	scan := normalize.Scan(raw)

	output := map[string]any{
		"id":          scan.ID,
		"status":      scan.Status,
		"queue_state": scan.QueueState,
		"message":     "Scan started or queued",
	}

	b, _ := json.Marshal(output)
	return string(b), nil
}

func (t *DASTScanStartTool) GetEnforcerProfile() framework.EnforcerProfile {
	return framework.NewEnforcerProfile(
		framework.WithRisk(framework.RiskHigh),
		framework.WithImpact(framework.ImpactWrite),
		framework.WithPII(false),
		framework.WithIdempotent(false),
		framework.WithApprovalReq(true),
		framework.WithResourceCost(10),
	)
}

type DASTScanFromTemplateTool struct {
	client *client.Client
	cfg    interface{ ReadOnly() bool }
}

func NewDASTScanFromTemplateTool(c *client.Client, cfg interface{ ReadOnly() bool }) *DASTScanFromTemplateTool {
	return &DASTScanFromTemplateTool{client: c, cfg: cfg}
}

func (t *DASTScanFromTemplateTool) Name() string { return "dast_scan_from_template" }
func (t *DASTScanFromTemplateTool) Description() string {
	return "Start a DAST scan using an uploaded scan or template file"
}

func (t *DASTScanFromTemplateTool) Schema() mcp.ToolInputSchema {
	return mcp.ToolInputSchema{
		Type: "object",
		Properties: map[string]interface{}{
			"application_id": map[string]interface{}{
				"type":        "string",
				"description": "Application ID",
			},
			"file_id": map[string]interface{}{
				"type":        "string",
				"description": "Uploaded file ID (.scant or .scan file)",
			},
			"scan_name": map[string]interface{}{
				"type":        "string",
				"description": "Optional scan name",
			},
		},
		Required: []string{"application_id", "file_id"},
	}
}

func (t *DASTScanFromTemplateTool) Handle(ctx context.Context, args map[string]interface{}) (string, error) {
	if readonly.IsReadOnly(t.cfg) {
		return "", fmt.Errorf("server is in readonly mode")
	}

	appID, _ := args["application_id"].(string)
	fileID, _ := args["file_id"].(string)
	scanName, _ := args["scan_name"].(string)

	if appID == "" || fileID == "" {
		return "", fmt.Errorf("application_id and file_id are required")
	}

	type ScanRequest struct {
		ApplicationID      string `json:"ApplicationId"`
		ScanOrTemplateFile string `json:"ScanOrTemplateFileId"`
		Name               string `json:"Name,omitempty"`
	}

	req := ScanRequest{
		ApplicationID:      appID,
		ScanOrTemplateFile: fileID,
	}
	if scanName != "" {
		req.Name = scanName
	}

	resp, err := t.client.Post("/scans/dast/fromfile", req)
	if err != nil {
		return "", fmt.Errorf("start scan from template: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		return "", client.ParseError(resp)
	}

	var raw map[string]any
	if err := client.DecodeJSON(resp, &raw); err != nil {
		return "", fmt.Errorf("decode response: %w", err)
	}

	scan := normalize.Scan(raw)

	output := map[string]any{
		"id":          scan.ID,
		"status":      scan.Status,
		"queue_state": scan.QueueState,
		"message":     "Scan started from template",
	}

	b, _ := json.Marshal(output)
	return string(b), nil
}

func (t *DASTScanFromTemplateTool) GetEnforcerProfile() framework.EnforcerProfile {
	return framework.NewEnforcerProfile(
		framework.WithRisk(framework.RiskHigh),
		framework.WithImpact(framework.ImpactWrite),
		framework.WithPII(false),
		framework.WithIdempotent(false),
		framework.WithApprovalReq(true),
		framework.WithResourceCost(10),
	)
}

type ScanCancelTool struct {
	client *client.Client
	cfg    interface{ ReadOnly() bool }
}

func NewScanCancelTool(c *client.Client, cfg interface{ ReadOnly() bool }) *ScanCancelTool {
	return &ScanCancelTool{client: c, cfg: cfg}
}

func (t *ScanCancelTool) Name() string        { return "scan_cancel" }
func (t *ScanCancelTool) Description() string { return "Cancel a queued or running scan" }

func (t *ScanCancelTool) Schema() mcp.ToolInputSchema {
	return mcp.ToolInputSchema{
		Type: "object",
		Properties: map[string]interface{}{
			"id": map[string]interface{}{
				"type":        "string",
				"description": "Scan ID to cancel",
			},
		},
		Required: []string{"id"},
	}
}

func (t *ScanCancelTool) Handle(ctx context.Context, args map[string]interface{}) (string, error) {
	if readonly.IsReadOnly(t.cfg) {
		return "", fmt.Errorf("server is in readonly mode")
	}

	id, _ := args["id"].(string)
	if id == "" {
		return "", fmt.Errorf("id is required")
	}

	path := fmt.Sprintf("/scans/%s/cancel", id)
	resp, err := t.client.Post(path, nil)
	if err != nil {
		return "", fmt.Errorf("cancel scan: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 && resp.StatusCode != 204 {
		return "", client.ParseError(resp)
	}

	output := map[string]any{
		"id":      id,
		"message": "Scan cancellation requested",
	}

	b, _ := json.Marshal(output)
	return string(b), nil
}

func (t *ScanCancelTool) GetEnforcerProfile() framework.EnforcerProfile {
	return framework.NewEnforcerProfile(
		framework.WithRisk(framework.RiskHigh),
		framework.WithImpact(framework.ImpactWrite),
		framework.WithPII(false),
		framework.WithIdempotent(false),
		framework.WithApprovalReq(true),
	)
}
