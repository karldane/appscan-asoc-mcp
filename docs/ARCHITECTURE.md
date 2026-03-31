# Architecture

## Overview

This document describes the internal architecture of the AppScan on Cloud MCP Server.

## High-Level Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                     MCP Bridge (External)                   │
│                   (adds appscan_ namespace)                 │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                    mcp-framework Server                     │
│  ┌─────────────────────────────────────────────────────────┐  │
│  │                    Tool Handlers                       │  │
│  │  apps_list, scans_list, findings_search, etc.        │  │
│  └─────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                     Internal Packages                        │
├──────────┬──────────┬──────────┬──────────┬────────────────┤
│  config  │ readonly │  client  │  model  │   normalize    │
│          │          │          │          │                │
└──────────┴──────────┴──────────┴──────────┴────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                  AppScan on Cloud API                       │
└─────────────────────────────────────────────────────────────┘
```

## Package Structure

### `main.go`
Entry point. Loads configuration, creates client, registers tools, starts server.

### `internal/config`
Configuration management:
- Reads from environment variables (`APPSCAN_BASE_URL`, `APPSCAN_KEY_ID`, etc.)
- CLI flag parsing with precedence over env vars
- Provides `ReadOnly()` method for readonly mode check

### `internal/readonly`
Readonly enforcement:
- `IsReadOnly(cfg)` - checks if config indicates readonly mode
- Each tool calls this check before executing mutating operations

### `internal/client`
ASoC HTTP client:
- Constructs authenticated requests (Basic auth)
- Provides GET/POST/PUT/DELETE methods
- Error handling and response decoding
- Custom request support for file uploads

### `internal/model`
Normalized data structures:
- `Application`, `Scan`, `Finding`, `Report`, `FileInfo`
- `AssetGroup`, `Policy`
- All include `Raw` field for source fidelity

### `internal/normalize`
API response mappers:
- Convert ASoC API responses to normalized model types
- Handle status/severity normalization
- Date parsing

### `internal/tools`
MCP tool implementations. Each tool implements `framework.ToolHandler`:
- `Name()` - tool name (without appscan_ prefix)
- `Description()` - user-facing description
- `Schema()` - JSON schema for parameters
- `Handle()` - execute the tool
- `GetEnforcerProfile()` - safety metadata

Tool groups:
- `apps.go` - Application tools
- `files.go` - File upload/get
- `scans.go` - Scan operations
- `findings.go` - Finding queries
- `reports.go` - Report generation
- `policies.go` - Asset groups, policies, compliance

## Data Flow

1. **Request received** → mcp-framework routes to tool handler
2. **Tool validates** → checks readonly mode, validates args
3. **Client executes** → makes HTTP request to ASoC
4. **Response normalized** → maps API response to model type
5. **Result returned** → JSON serialized for MCP response

## Safety Metadata

Each tool implements `GetEnforcerProfile()` returning:
- `RiskLevel` - low/med/high/critical
- `ImpactScope` - read/write/delete/admin
- `PIIExposure` - whether tool exposes sensitive data
- `Idempotent` - whether safe to retry
- `ApprovalReq` - whether human approval needed
- `ResourceCost` - CPU/API credit weight (1-10)

This metadata is transmitted during `tools/list` via tool annotations.

## Configuration Precedence

1. CLI flags (highest)
2. Environment variables
3. Default values

Example:
```
APPSCAN_BASE_URL=https://api.example.com ./appscan-asoc-mcp --base-url=https://override.com
# Result: base-url is https://override.com
```

## Testing Strategy

- Unit tests for config, readonly, client, normalize
- Build verification for tools
- Full test suite via `make test`

## Extension Points

The architecture supports future expansion:
- Add new tool groups (SAST, SCA, IAST) in `internal/tools/`
- Normalization layer handles new entity types
- Client provides reusable HTTP patterns