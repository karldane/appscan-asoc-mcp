package main

import (
	"fmt"
	"os"

	"appscan-asoc-mcp/internal/client"
	"appscan-asoc-mcp/internal/config"
	"appscan-asoc-mcp/internal/tools"

	"github.com/karldane/mcp-framework/framework"
)

func main() {
	cfg := config.Load()
	cfg = config.MergeCLIFlags(cfg)

	if cfg.BaseURL == "" || cfg.KeyID == "" || cfg.KeySecret == "" {
		fmt.Fprintln(os.Stderr, "Error: APPSCAN_BASE_URL, APPSCAN_KEY_ID (or APPSCAN_API_KEY), and APPSCAN_KEY_SECRET are required")
		os.Exit(1)
	}

	asocClient := client.New(cfg.BaseURL, cfg.KeyID, cfg.KeySecret, cfg.TimeoutSeconds)

	serverConfig := &framework.Config{
		Name:    "appscan-asoc-mcp",
		Version: "1.0.0",
		Instructions: `AppScan on Cloud MCP Server

This server provides tools for interacting with HCL AppScan on Cloud.

Required environment variables:
- APPSCAN_BASE_URL: Base URL for the ASoC API (e.g., https://cloud.appscan.com/api/v4)
- APPSCAN_KEY_ID: API key ID
- APPSCAN_KEY_SECRET: API key secret
- Or use APPSCAN_API_KEY=keyid:keysecret for combined form`,
	}

	server := framework.NewServerWithConfig(serverConfig)
	server.SetWriteEnabled(!cfg.ReadOnly())

	server.RegisterTool(tools.NewAppsListTool(asocClient, cfg))
	server.RegisterTool(tools.NewAppGetTool(asocClient, cfg))
	server.RegisterTool(tools.NewAppsSearchTool(asocClient, cfg))
	server.RegisterTool(tools.NewFilesUploadTool(asocClient, cfg))
	server.RegisterTool(tools.NewFileGetTool(asocClient, cfg))
	server.RegisterTool(tools.NewScansListTool(asocClient, cfg))
	server.RegisterTool(tools.NewScanGetTool(asocClient, cfg))
	server.RegisterTool(tools.NewScanStatusTool(asocClient, cfg))
	server.RegisterTool(tools.NewDASTScanStartTool(asocClient, cfg))
	server.RegisterTool(tools.NewDASTScanFromTemplateTool(asocClient, cfg))
	server.RegisterTool(tools.NewScanCancelTool(asocClient, cfg))
	server.RegisterTool(tools.NewFindingsListTool(asocClient, cfg))
	server.RegisterTool(tools.NewFindingsSearchTool(asocClient, cfg))
	server.RegisterTool(tools.NewFindingGetTool(asocClient, cfg))
	server.RegisterTool(tools.NewFindingGroupSummaryTool(asocClient, cfg))
	server.RegisterTool(tools.NewReportsListTool(asocClient, cfg))
	server.RegisterTool(tools.NewReportGenerateTool(asocClient, cfg))
	server.RegisterTool(tools.NewReportGetTool(asocClient, cfg))
	server.RegisterTool(tools.NewAssetGroupsListTool(asocClient, cfg))
	server.RegisterTool(tools.NewPoliciesListTool(asocClient, cfg))
	server.RegisterTool(tools.NewComplianceSummaryTool(asocClient, cfg))

	server.Initialize()
	fmt.Fprintf(os.Stderr, "AppScan ASoC MCP Server initialized\n")
	fmt.Fprintf(os.Stderr, "Read-only mode: %v\n", cfg.ReadOnly())
	fmt.Fprintf(os.Stderr, "Tools: %v\n", server.ListTools())

	if err := server.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
		os.Exit(1)
	}
}
