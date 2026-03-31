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
export APPSCAN_BASE_URL="https://cloud.appscan.com/api/v4"
export APPSCAN_KEY_ID="your-key-id"
export APPSCAN_KEY_SECRET="your-key-secret"
```

Or use combined API key form:
```bash
export APPSCAN_BASE_URL="https://cloud.appscan.com/api/v4"
export APPSCAN_API_KEY="keyid:keysecret"
```

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

## Architecture

See [ARCHITECTURE.md](docs/ARCHITECTURE.md) for detailed design documentation.

## Testing

```bash
make test
```

Run specific test packages:
```bash
go test ./internal/config/... -v
go test ./internal/normalize/... -v
```

## License

See LICENSE file.