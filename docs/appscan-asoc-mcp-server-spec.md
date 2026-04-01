# Specification: AppScan on Cloud MCP Server (Go)

## Overview

This document defines the specification for a Go-based MCP server that integrates with **HCL AppScan on Cloud (ASoC)** only. The server is intended to run as a backend subprocess behind `mcp-bridge`, and it will be built using the `mcp-framework` so that every tool self-reports its safety profile through tool annotations during `tools/list`.[cite:151][cite:157][cite:123][cite:122]

The primary design goal is to cover at least **80% of day-to-day developer-useful ASoC functionality**, with emphasis on DAST because that is the currently available product for testing. The server should normalize data models where practical so that future SAST, SCA, and IAST support can fit the same tool contracts, while allowing DAST-first implementation in v1.[cite:164][cite:173][cite:123]

The server must support both **read** and **write** operations, but it must also provide a `--readonly` startup switch that disables any mutating action at runtime even if the tool is present in the tool list. This is in addition to the bridge-side policy enforcement enabled by the framework’s self-reporting safety metadata.[cite:122][cite:123]

## Scope

### In scope

The server will target the **AppScan on Cloud REST API v4** only. Authentication must use AppScan’s API-key based authentication model, which the public documentation describes as either direct `X-Api-Key` usage or token-based authentication flows for API access.[cite:151][cite:157]

The required implementation scope is:
- Applications and application metadata.
- DAST scan creation, launch, listing, status, and scan detail retrieval.
- Findings and issue exploration for day-to-day remediation workflows.
- Report generation and report retrieval metadata.
- File upload workflows for scan files and templates, including DAST template-driven and report-only flows where supported by ASoC.
- Lightweight policy/compliance and asset-group reads when they materially affect developer workflows.[cite:157][cite:164][cite:173][cite:180]

### Out of scope for v1

The following should be marked **optional** or deferred because they are not core developer workflows or are primarily administrative:
- User and organization administration.
- Audit and tenant administration features.
- Advanced integration administration and webhook configuration.
- Broad support for non-DAST scan families beyond normalized placeholders and shared domain models.
- Any AppScan Enterprise or non-ASoC product surface.[cite:151][cite:173]

## Design principles

The implementation must follow these principles:
- **DAST-first**: optimize for the functionality that can actually be tested and used now.[cite:164]
- **Normalized outputs**: where ASoC returns product-specific payloads, map them into stable, MCP-friendly result shapes with a `raw` field for source fidelity.[cite:173]
- **Framework-native safety**: every tool must implement `GetEnforcerProfile()` and emit safety metadata via annotations as described in the attached framework docs.[cite:122][cite:123]
- **Bridge-friendly naming**: tool names should omit the `appscan_` prefix because the bridge adds that namespace itself.
- **Read-only guardrail**: `--readonly` must block all mutating tools at runtime with a clear error, even if the tool remains discoverable for compatibility and policy visibility.
- **Subprocess simplicity**: configure entirely through environment variables and CLI flags so the server works cleanly in `mcp-bridge` per-user process pools.[cite:4]

## Runtime configuration

### Required environment variables

The server should support these environment variables:

| Variable | Required | Purpose |
|---|---|---|
| `APPSCAN_BASE_URL` | Yes | Hostname of the ASoC tenant (e.g. `https://eu.cloud.appscan.com`). The server appends `/api/v4` automatically and strips any existing path suffix. Regional values: EU=`https://eu.cloud.appscan.com`, US=`https://cloud.appscan.com`. |
| `APPSCAN_KEY_ID` | Yes | AppScan API key ID. |
| `APPSCAN_KEY_SECRET` | Yes | AppScan API key secret. |
| `APPSCAN_API_KEY` | Optional | Combined `KeyID:KeySecret` form if provided directly. Takes precedence over individual key vars. |
| `APPSCAN_TIMEOUT_SECONDS` | Optional | Default outbound HTTP timeout. |
| `APPSCAN_POLL_INTERVAL_SECONDS` | Optional | Poll interval for long-running operations if polling helpers are implemented. |

These settings match the bridge’s model of per-user credential injection at process spawn time.[cite:4][cite:151][cite:157]

### Required CLI flags

| Flag | Required | Purpose |
|---|---|---|
| `--readonly` | No | Disable execution of all mutating tools. |
| `--log-json` | No | Emit structured logs for bridge aggregation. |
| `--timeout` | No | Override default HTTP timeout. |
| `--base-url` | No | Override `APPSCAN_BASE_URL`. |

If both CLI flags and environment variables are present, CLI flags should take precedence.[cite:4]

## Tool groups

The backend tool names below are intentionally unprefixed. The bridge is expected to add the `appscan_` namespace when exposing them to clients.

### Required tools

#### Applications

| Tool | Purpose | Mutating | Priority |
|---|---|---|---|
| `apps_list` | List applications with filtering and pagination. | No | Required |
| `app_get` | Fetch detailed application metadata. | No | Required |
| `apps_search` | Search applications by name, tag, business unit, or status. | No | Required |
| `app_create` | Create an application with basic metadata. | Yes | Required |
| `app_update` | Update application metadata and selected settings. | Yes | Required |

Applications are a core ASoC organizational object and must be first-class because all scan and findings workflows anchor to an application.[cite:151][cite:173]

#### File upload and template workflows

| Tool | Purpose | Mutating | Priority |
|---|---|---|---|
| `files_upload` | Upload a scan/template/config file and return file metadata including file ID. | Yes | Required |
| `file_get` | Retrieve file metadata by ID if the API exposes it. | No | Required |
| `dast_scan_from_template` | Start a DAST scan using an uploaded `.scant` or `.scan` file. | Yes | Required |
| `dast_report_from_template` | Trigger report-only processing from an uploaded scan/template file where supported. | Yes | Required |

This group is mandatory because ASoC documentation explicitly describes file-upload driven DAST and report-only flows using uploaded AppScan Standard scan or template files and a `FileId`/`ScanOrTemplateFileId` workflow.[cite:157][cite:164]

#### Scans

| Tool | Purpose | Mutating | Priority |
|---|---|---|---|
| `scans_list` | List scans with filters by app, family, state, and date. | No | Required |
| `scan_get` | Get detailed scan metadata. | No | Required |
| `scan_status` | Return normalized scan state and queue/execution details. | No | Required |
| `dast_scan_start` | Launch a DAST scan for an application or target. | Yes | Required |
| `scan_rescan` | Re-run or resubmit a prior scan where API support exists. | Yes | Required |
| `scan_cancel` | Cancel a queued or running scan if supported by API permissions. | Yes | Required |

Because the tenant has only one licensed runner, scans may be queued rather than started immediately. The server should treat “queued” and “started” as normal outcomes and return them clearly without adding complicated wait-mode behavior.[cite:173]

#### Findings and remediation

| Tool | Purpose | Mutating | Priority |
|---|---|---|---|
| `findings_list` | List findings/issues for an application or scan. | No | Required |
| `findings_search` | Search and filter findings by severity, status, issue type, compliance, or text. | No | Required |
| `finding_get` | Return detailed issue data with normalized fields and raw payload. | No | Required |
| `finding_group_summary` | Aggregate findings by severity, issue type, status, or compliance category. | No | Required |

The ASoC results experience is centered on finding review, filtering, and remediation workflows, so these are core developer tools rather than optional extras.[cite:173][cite:180]

#### Reports

| Tool | Purpose | Mutating | Priority |
|---|---|---|---|
| `reports_list` | List reports or report definitions available for an app or scan. | No | Required |
| `report_generate` | Request report generation for an application or scan. | Yes | Required |
| `report_get` | Retrieve report status and metadata, including downloadable reference if available. | No | Required |

Reports are a regular operational need for validation, handoff, and governance, and ASoC docs explicitly document report generation and sample report workflows.[cite:158][cite:173]

#### Policy and grouping reads

| Tool | Purpose | Mutating | Priority |
|---|---|---|---|
| `asset_groups_list` | List asset groups relevant to application placement and filtering. | No | Required |
| `policies_list` | List policies or policy-related metadata relevant to scans and compliance. | No | Required |
| `compliance_summary` | Return normalized compliance/policy summary for an app or scan. | No | Required |

These are included as read tools because they help developers understand policy failures and scan placement without drifting into heavy admin territory.[cite:160][cite:173]

### Optional tools

| Tool | Purpose | Mutating | Reason optional |
|---|---|---|---|
| `asset_group_get` | Retrieve one asset group in detail. | No | Lower day-to-day frequency. |
| `asset_group_assign` | Assign an app to an asset group. | Yes | More administrative than developer-centric. |
| `report_download_binary` | Return binary report contents or a file artifact path. | No | Useful, but can be deferred if metadata plus download reference is enough. |
| `scan_delete` | Delete historical scans or cleanup artifacts. | Yes | Operational, not day-to-day. |
| `users_list` | List users. | No | Admin-only. |
| `user_get` | Retrieve user details. | No | Admin-only. |
| `org_settings_get` | Inspect organization settings. | No | Admin-only. |
| `org_settings_update` | Modify org settings. | Yes | Admin-only and risky. |

## Normalized data contracts

### Application summary

Every application-returning tool should normalize to:

```json
{
  "id": "string",
  "name": "string",
  "description": "string|null",
  "business_unit": "string|null",
  "tags": ["string"],
  "asset_group_ids": ["string"],
  "created_at": "RFC3339|null",
  "updated_at": "RFC3339|null",
  "raw": {}
}
```

### Scan summary

Every scan-returning tool should normalize to:

```json
{
  "id": "string",
  "app_id": "string|null",
  "scan_family": "dast",
  "status": "queued|running|completed|failed|canceled|unknown",
  "queue_state": "queued|not_queued|unknown",
  "target": "string|null",
  "submitted_at": "RFC3339|null",
  "started_at": "RFC3339|null",
  "completed_at": "RFC3339|null",
  "is_readonly_safe": false,
  "raw": {}
}
```

### Finding summary

Every findings tool should normalize to:

```json
{
  "id": "string",
  "app_id": "string|null",
  "scan_id": "string|null",
  "scan_family": "dast",
  "title": "string",
  "severity": "info|low|medium|high|critical|unknown",
  "status": "open|fixed|new|ignored|noise|unknown",
  "issue_type": "string|null",
  "location": "string|null",
  "first_seen": "RFC3339|null",
  "last_seen": "RFC3339|null",
  "compliance": ["string"],
  "recommendation": "string|null",
  "raw": {}
}
```

### Report summary

Every report-returning tool should normalize to:

```json
{
  "id": "string",
  "app_id": "string|null",
  "scan_id": "string|null",
  "status": "pending|ready|failed|unknown",
  "format": "pdf|html|xml|json|unknown",
  "download_url": "string|null",
  "created_at": "RFC3339|null",
  "raw": {}
}
```

The `raw` field is mandatory in all normalized outputs so that no source fidelity is lost while the model-facing result stays stable.[cite:173]

## Readonly behavior

The server must support a `--readonly` mode with the following rules:
- Any mutating tool remains discoverable in `tools/list` so clients and bridges can understand the full capability surface.
- Mutating tools must fail fast at execution time with a deterministic error such as `server is in readonly mode`.
- The tool annotations should still accurately describe the tool as write/admin class even when readonly mode is enabled.
- Optional enhancement: add a `_meta.readonly_blocked=true` hint in tool annotations or tool descriptions during startup.

This preserves discoverability while preventing accidental writes and keeping the bridge’s safety logic accurate.[cite:122][cite:123]

## Safety profiles

Each tool must implement `GetEnforcerProfile()` using the framework. The minimum classification guidance is:

| Tool category | Risk | Impact | PII | Idempotent | Approval default |
|---|---|---|---|---|---|
| Application reads | Med | Read | True | True | False |
| Findings reads | Med | Read | True | True | False |
| Report metadata reads | Med | Read | True | True | False |
| Asset/policy reads | Low-Med | Read | False/True | True | False |
| App create/update | High | Write | True | False | False |
| File upload | High | Write | True | False | True |
| Scan launch/rescan/cancel | High | Write | True | False | True |
| Org/user admin | Critical | Admin | True | False | True |

The exact thresholds can be adjusted, but the above should be the default baseline for v1 because AppScan findings often include sensitive URLs, parameters, and remediation context, while uploads and scans consume scarce runner capacity and can materially affect tenant state.[cite:122][cite:173]

## Error handling

The server should normalize common failure modes into structured MCP-friendly errors:
- Authentication failure.
- Permission denied.
- Resource not found.
- Validation error.
- Rate-limit or concurrency limit encountered.
- Upstream queue state returned instead of started state.
- Readonly mode violation.
- Unsupported operation for current product/license tier.

Whenever possible, the server should preserve upstream AppScan error bodies under a structured `details.raw_error` field while presenting a stable message to the caller.[cite:123]

## Implementation architecture

### Packages

Suggested package structure:

```text
appscan-mcp/
├── main.go
├── internal/
│   ├── config/
│   ├── client/
│   ├── model/
│   ├── normalize/
│   ├── tools/
│   │   ├── apps.go
│   │   ├── files.go
│   │   ├── scans.go
│   │   ├── findings.go
│   │   ├── reports.go
│   │   └── policies.go
│   └── readonly/
├── go.mod
├── README.md
└── LICENSE
```

### Client layer

The client layer should:
- Construct authenticated HTTP requests.
- Support both direct `X-Api-Key` composition and any documented token exchange flow if needed later.[cite:151][cite:157]
- Apply consistent timeout, retry, and pagination helpers.
- Log request IDs and response codes without leaking secrets.

### Tool layer

Each tool should be a small focused handler implementing the framework’s `ToolHandler` interface, with schema validation, normalized mapping, and `GetEnforcerProfile()` defined alongside the handler.[cite:123]

### Normalization layer

Normalization should be centralized so that scan, finding, and report mappers can be reused if SAST, SCA, or IAST support is added later. DAST-first behavior should not hardcode DAST assumptions into the core result structs beyond the `scan_family` field default.[cite:173]

## Testing strategy

The server should be developed with TDD in line with the framework’s guidance.[cite:123]

### Required tests
- Config parsing tests.
- Readonly enforcement tests.
- Client authentication header tests.
- Normalization tests for apps, scans, findings, and reports.
- Tool schema tests.
- Safety annotation tests that verify every tool returns the expected `EnforcerProfile`.
- Integration tests against mocked ASoC HTTP responses.
- Optional live smoke tests gated by environment variables against a real tenant.

### Live test coverage priorities

Because only ASoC DAST is currently available, live integration testing should prioritize:
1. `apps_list`, `app_get`.
2. `files_upload` using a valid template or scan file.
3. `dast_scan_from_template` and `dast_scan_start`.
4. `scans_list`, `scan_status`, `scan_get`.
5. `findings_list`, `finding_get`.
6. `report_generate`, `report_get`.[cite:164][cite:173]

## Minimum viable release definition

Version 1.0 of the server is complete when all of the following are true:
- The server authenticates against ASoC successfully using API-key credentials.[cite:151][cite:157]
- Required tools listed in this spec are implemented and exposed through the framework.[cite:123]
- Every tool self-reports safety metadata through `GetEnforcerProfile()`.[cite:122][cite:123]
- `--readonly` blocks all mutating tool execution reliably.
- DAST file upload plus template/scan-driven launch flow is implemented and tested.[cite:157][cite:164]
- Findings and report workflows are normalized and usable for real remediation and verification work.[cite:173][cite:180]

## Recommended follow-up phases

### Phase 2
- Add binary report download helper.
- Add richer compliance and policy detail tools.
- Add pagination cursors and advanced filtering helpers on large finding sets.
- Add better queue visibility if ASoC exposes runner/queue metadata.[cite:173]

### Phase 3
- Expand the normalized model to SAST, SCA, and IAST when test access exists.
- Add optional administrative tools under explicit feature flags.
- Add cached lookup helpers for applications, asset groups, and policy metadata.

## Implementation notes (v1 discoveries)

The following were confirmed against the real EU ASoC tenant during v1 implementation and take precedence over any contradictory assumptions in this spec:

### Authentication

ASoC v4 uses **direct API key authentication** via the `X-Api-Key` request header with value `KeyID:KeySecret`. No token exchange or Bearer token is required for API calls. Attempting Bearer token auth returns 401.

### API endpoint casing

The ASoC v4 API uses PascalCase paths, not lowercase:
- Applications: `GET /api/v4/Apps` (not `/applications`)
- List responses use an `Items` key (not a resource-named key like `Applications`)
- Single resource: `GET /api/v4/Apps/{id}`

All other endpoints (Scans, Findings, Reports, etc.) also use PascalCase and follow the same `Items`-based list response shape.

### Base URL handling

The server normalizes `APPSCAN_BASE_URL` at startup: if the value does not end with `/api/v4`, the server appends it. This means the caller should pass the hostname only (e.g. `https://eu.cloud.appscan.com`). Tool path strings are written relative to the v4 root (e.g. `/Apps`, `/scans/dast`) without any `/api/v4` prefix - the client layer concatenates the full URL.

### Mutating tools in readonly mode

Resolved: mutating tools remain visible in `tools/list` but fail at execution time with a clear `"server is in readonly mode"` error. No description suffix is added.

## Open decisions

The following decisions are intentionally left for implementation kickoff:
- Whether report binary download should be included in v1 or deferred to v1.1.
- Which exact upload MIME types and file-size limits should be validated client-side versus relying on ASoC upstream validation.
- Which policy/compliance endpoints are easy enough to justify in v1 once the real tenant surface is enumerated.

## Summary

This server should be implemented as a **DAST-first, ASoC-only, Go MCP backend** using `mcp-framework`, with normalized outputs, strong safety self-reporting, runtime readonly enforcement, and first-class support for file-upload-driven DAST scan workflows. That scope should cover the majority of practical developer interactions with AppScan on Cloud while leaving administrative and low-frequency APIs explicitly optional.[cite:164][cite:173][cite:122][cite:123]
