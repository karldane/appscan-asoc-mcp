# AppScan on Cloud MCP Server

A Go-based MCP (Model Context Protocol) server for **HCL AppScan on Cloud (ASoC)** using `mcp-framework`. Provides tools for interacting with AppScan on Cloud's DAST capabilities.

## Features

- **Applications** - List, get, search applications
- **Files** - Upload scan/template files, get file metadata
- **Scans** - Start DAST scans, list scans, get status, cancel scans
- **Findings** - List, search, get details, group summaries
- **Reports** - Generate, list, retrieve report metadata
- **Policies** - List asset groups, policies, compliance summaries

## Safety Features

- Every tool self-reports safety metadata via `GetEnforcerProfile()`
- `--readonly` flag to disable all mutating tools
- Tools remain visible but fail at execution time when readonly

## Requirements

- Go 1.25+
- AppScan on Cloud API credentials

## Installation

```bash
make build
```

Or install to GOPATH:
```bash
make install
```

## Usage

Set required environment variables:

```bash
export APPSCAN_BASE_URL="https://eu.cloud.appscan.com"
export APPSCAN_KEY_ID="your-key-id"
export APPSCAN_KEY_SECRET="your-key-secret"
```

Or use combined API key form:
```bash
export APPSCAN_BASE_URL="https://eu.cloud.appscan.com"
export APPSCAN_API_KEY="keyid:keysecret"
```

> **Note:** `APPSCAN_BASE_URL` should be the **hostname only** (e.g. `https://eu.cloud.appscan.com`). The server automatically appends `/api/v4` and strips any existing path suffix to avoid double-pathing. If you pass a URL that already ends with `/api/v4` it is also handled correctly.

Run the server:
```bash
./appscan-asoc-mcp
```

### CLI Flags

| Flag | Description |
|------|-------------|
| `--readonly` | Disable all mutating tools |
| `--base-url` | Override APPSCAN_BASE_URL |
| `--timeout` | HTTP timeout in seconds |
| `--log-json` | Emit structured JSON logs |

## Available Tools

See [Tool Reference](docs/TOOLS.md) for detailed tool documentation.

### Application Tools
- `apps_list` - List applications with pagination
- `app_get` - Get application details by ID
- `apps_search` - Search applications

### File Tools
- `files_upload` - Upload scan/template files
- `file_get` - Get file metadata

### Scan Tools
- `scans_list` - List scans with filters
- `scan_get` - Get scan details
- `scan_status` - Get normalized scan state
- `dast_scan_start` - Start a DAST scan
- `dast_scan_from_template` - Start scan from uploaded file
- `scan_cancel` - Cancel a scan

### Findings Tools
- `findings_list` - List findings
- `findings_search` - Search/filter findings
- `finding_get` - Get finding details
- `finding_group_summary` - Aggregate findings

### Report Tools
- `reports_list` - List reports
- `report_generate` - Generate a report
- `report_get` - Get report status/metadata

### Policy Tools
- `asset_groups_list` - List asset groups
- `policies_list` - List security policies
- `compliance_summary` - Get compliance summary

## Authentication

The server authenticates to ASoC using the `X-Api-Key` request header with the value `KeyID:KeySecret`. This is the direct API key form - no token exchange is required. The header is set on every outgoing HTTP request by the client layer.

> **Regional endpoints:** For EU tenants use `https://eu.cloud.appscan.com`. For US (default) use `https://cloud.appscan.com`.

## ASoC API Notes

- All API calls target **REST API v4** (`/api/v4/...`). The base URL should be the hostname only; the server handles the `/api/v4` prefix internally.
- The applications endpoint is `/api/v4/Apps` (capitalized, not `/applications`). List responses use an `Items` key. This reflects the actual ASoC v4 API surface, which differs from some documentation examples.
- Scan results may be `queued` rather than `started` when the tenant has only one runner - this is a valid outcome, not an error.

## Architecture

See [ARCHITECTURE.md](docs/ARCHITECTURE.md) for detailed design documentation.

## Testing

```bash
make test
```

Run specific test packages:
```bash
go test ./internal/tools/... -v
go test ./internal/normalize/... -v
go test ./internal/config/... -v
```

Coverage summary (as of v1.0):

| Package | Coverage |
|---|---|
| `internal/tools` | ~86% |
| `internal/normalize` | ~96% |
| `internal/readonly` | 100% |
| `internal/config` | ~68% |

## License

See LICENSE file.