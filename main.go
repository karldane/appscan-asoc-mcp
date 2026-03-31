package main

import (
	"context"
	"fmt"
	"os"

	"github.com/karldane/mcp-framework/framework"
	"github.com/mark3labs/mcp-go/mcp"
)

type HelloTool struct{}

func (t *HelloTool) Name() string        { return "hello" }
func (t *HelloTool) Description() string { return "Say hello to someone" }

func (t *HelloTool) Schema() mcp.ToolInputSchema {
	return mcp.ToolInputSchema{
		Type: "object",
		Properties: map[string]interface{}{
			"name": map[string]interface{}{
				"type":        "string",
				"description": "Name of the person to greet",
			},
		},
		Required: []string{"name"},
	}
}

func (t *HelloTool) Handle(ctx context.Context, args map[string]interface{}) (string, error) {
	name, ok := args["name"].(string)
	if !ok || name == "" {
		return "", fmt.Errorf("name is required")
	}
	return fmt.Sprintf("Hello, %s!", name), nil
}

func (t *HelloTool) GetEnforcerProfile() *framework.EnforcerProfile {
	return framework.DefaultEnforcerProfile()
}

func main() {
	config := &framework.Config{
		Name:    "appscan-asoc-mcp",
		Version: "1.0.0",
		Instructions: `AppScan on Cloud MCP Server

This server provides tools for interacting with HCL AppScan on Cloud.`,
	}

	server := framework.NewServerWithConfig(config)

	if err := server.RegisterTool(&HelloTool{}); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to register hello tool: %v\n", err)
		os.Exit(1)
	}

	server.Initialize()
	fmt.Fprintln(os.Stderr, "AppScan ASoC MCP Server initialized")
	fmt.Fprintln(os.Stderr, "Tools:", server.ListTools())

	if err := server.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
		os.Exit(1)
	}
}
