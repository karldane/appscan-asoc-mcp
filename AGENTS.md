# AGENTS.md - Developer Handbook

**Project:** appscan-asoc-mcp-server  
**Language:** Go 1.25.x preferred, standard library first  
**Primary spec:** `docs/appscan-asoc-mcp-server-spec.md`

## Mission

Build a Go-based MCP server for **HCL AppScan on Cloud (ASoC)** using `mcp-framework`, with a **DAST-first** implementation that covers the majority of day-to-day developer workflows while remaining extensible for future scan families.

This server is intended to run behind an MCP bridge that adds the external `appscan_` namespace prefix. The backend server itself should expose clean, unprefixed tool names such as `apps_list`, `dast_scan_start`, and `findings_search`.

## Source of Truth

Always follow these documents in order:

1. **`docs/appscan-asoc-mcp-server-spec.md`** - Primary functional and architectural specification.
2. **`README.md`** - Repository-level overview and setup instructions.
3. **`MCP-safety-reporting-spec.md`** - Safety metadata contract and bridge interpretation model.
4. **`mcp-framework` docs / code** - Framework usage patterns and interfaces.

If implementation details conflict with assumptions, update the spec first or record the discrepancy explicitly before proceeding.

## Go version policy

Use **Go 1.25.x** as the development and minimum supported target. This is the recommended cutoff because it is a current, supported Go release and matches the local development environment already available for this project.[cite:202][cite:203][cite:209][cite:212]

Rules:
- Set the module `go` version to `1.25`.
- Develop and test primarily on Go 1.25.8.
- Prefer stable standard-library features available in Go 1.25.
- Do not adopt 1.26-only features unless the project support policy is explicitly revised.
- Choose clarity and idiomatic Go over clever language-feature usage.

## Non-Negotiables

### 1. TDD always

This project uses **TDD without exception**.

Rules:
- No production code without a failing test first.
- Follow **red -> green -> refactor** for every feature, bug fix, and regression.
- Write the smallest failing test that proves the next increment of behavior.
- Refactor only after tests are green.
- If a bug is found, first write a regression test that fails for the bug.

Minimum practice:
- Unit tests for normalization, config parsing, readonly enforcement, tool schemas, and safety annotations.
- HTTP client tests with mocked ASoC responses.
- Integration-style tests at the tool-handler level.
- Optional live smoke tests only behind environment-gated flags.

### 2. Safety metadata is mandatory

Every tool must implement `GetEnforcerProfile()` correctly. There are **no exceptions**.

This server is designed to work with bridge-side policy enforcement. Tool safety metadata is a functional requirement, not a documentation nicety.

At minimum, every tool must accurately self-report:
- Risk level.
- Impact scope.
- Resource cost.
- PII exposure.
- Idempotence.
- Whether human approval is required.

If a tool mutates tenant state, consumes runner capacity, uploads files, starts or cancels scans, or performs admin actions, its profile must reflect that honestly.

### 3. Readonly mode must be real

`--readonly` is not cosmetic.

Requirements:
- Mutating tools may remain visible in `tools/list`.
- Mutating tools must fail deterministically at runtime when readonly mode is enabled.
- Tests must prove readonly enforcement for every mutating tool category.
- Tool annotations must still describe the true write/admin nature of the tool even when readonly mode is active.

### 4. Git checkpoints are required

Commit regularly at sensible checkpoints.

Rules:
- Commit after each logical slice reaches green tests.
- Do not batch unrelated work into one commit.
- Commit messages should explain **why** the change exists, not only what changed.
- Never leave large uncommitted working states unless actively mid-red phase.

Good checkpoint examples:
- `test: pin down config precedence and readonly parsing`
- `feat: add authenticated asoc client with api-key header support`
- `feat: implement apps_list and app_get with normalized mapping`
- `test: add regression coverage for findings severity normalization`
- `feat: add template upload and dast scan launch flows`

### 5. Subtask when useful

Use subtasking whenever the work naturally splits into independent areas.

Suggested subtask boundaries:
- Config and startup.
- Authenticated ASoC client.
- Normalization models.
- Applications tools.
- Files/upload tools.
- Scan tools.
- Findings tools.
- Report tools.
- Policy/compliance reads.
- Readonly enforcement.
- Safety annotation verification.
- Live smoke test harness.

Subtasks should produce concrete artifacts: code, tests, docs, or decision notes.

## Engineering Guidelines

### Go style

- Prefer the standard library whenever possible.
- Keep dependencies minimal and justified.
- Favor clear packages with narrow responsibilities.
- Avoid clever abstractions early; duplication is acceptable until patterns stabilize.
- Return structured errors with stable messages for MCP callers.
- Do not log secrets, tokens, headers, or sensitive finding content unnecessarily.

### Suggested package layout

```text
appscan-mcp/
├── main.go
├── internal/
│   ├── config/
│   ├── client/
│   ├── model/
│   ├── normalize/
│   ├── readonly/
│   ├── tools/
│   │   ├── apps.go
│   │   ├── files.go
│   │   ├── scans.go
│   │   ├── findings.go
│   │   ├── reports.go
│   │   └── policies.go
│   └── testutil/
├── docs/
│   └── appscan-asoc-mcp-server-spec.md
├── README.md
└── AGENTS.md
```

### Configuration

Support both env vars and flags, with flags taking precedence.

Expected configuration includes:
- `APPSCAN_BASE_URL`
- `APPSCAN_KEY_ID`
- `APPSCAN_KEY_SECRET`
- optional combined API key form
- timeout settings
- `--readonly`

Configuration parsing must be fully test-driven.

### HTTP client behavior

The ASoC client layer should:
- centralize auth header construction,
- centralize retries/timeouts,
- centralize pagination helpers where needed,
- preserve upstream error details for debugging,
- normalize status handling for queued/running/completed outcomes.

Test all of this with `httptest` servers before wiring real tools.

## Report handling

Version 1 should support **both** report metadata workflows and binary report retrieval if the implementation is not excessively complicated.[cite:158][cite:173]

Practical rule:
- `report_generate`, `report_get`, and metadata/list flows are mandatory for v1.
- Binary retrieval should be implemented in v1 if the upstream API makes it straightforward.
- If binary retrieval is deferred during implementation, record the reason explicitly in the spec or README rather than silently omitting it.

## Optional admin tools

Optional admin tools may be included in the codebase and tool list, but they must be self-reported accurately and classified conservatively.[cite:122][cite:123]

Rules:
- Admin and user-management tools are never to be disguised as ordinary read tools.
- Default safety posture for admin tools should be `risk=critical`, `impact=admin`, and typically `approval_req=true` unless a narrower case is justified.
- Admin tools must still honor `--readonly` and fail deterministically when writes are blocked.

## Tooling Guidelines

### Naming

Do **not** prefix tool names with `appscan_`. The bridge applies that namespace externally.

Use concise, domain-oriented names such as:
- `apps_list`
- `app_get`
- `files_upload`
- `dast_scan_start`
- `dast_scan_from_template`
- `findings_search`
- `report_generate`

### Required v1 priorities

The first implementation wave should focus on the most developer-relevant tools from the spec:
1. Config + startup + framework wiring.
2. Authenticated ASoC client.
3. Application list/get/search.
4. File upload.
5. DAST scan start / template-driven scan start.
6. Scan list/get/status.
7. Findings list/search/get.
8. Report generate/get/list.
9. Readonly enforcement across all mutating tools.
10. Safety annotation verification tests.

### Normalization

Normalize all major entities into stable model-facing shapes:
- application
- scan
- finding
- report

Always preserve a `raw` field so source fidelity is not lost.

Normalization code must be tested independently from tool handlers.

## Testing Strategy

### Standard commands

Run frequently:

```bash
go test ./... -count=1
```

Use targeted package runs while iterating, then run the full suite before every commit.

### Required test categories

- Config parsing and precedence.
- Readonly enforcement.
- Auth header generation.
- Client error mapping.
- Pagination helpers.
- Entity normalization.
- Tool input validation.
- Tool output mapping.
- Safety annotation correctness.
- Regression tests for bugs.

### Live tests

Live ASoC tests are allowed only when explicitly gated, for example via env vars such as:

```bash
APPSCAN_LIVE_TEST=1
APPSCAN_BASE_URL=...
APPSCAN_KEY_ID=...
APPSCAN_KEY_SECRET=...
go test ./... -run Live -count=1
```

Live tests must:
- be opt-in,
- skip cleanly when env vars are absent,
- avoid destructive behavior unless explicitly intended,
- respect the single-runner limitation,
- avoid long waits where queued status is a valid success outcome.

## Single Runner Constraint

The current tenant license provides **one runner only**.

Implications:
- A scan request may legitimately return **queued** instead of **started**.
- Tool results must treat queued as a valid outcome, not an error.
- Tests and tool descriptions should reflect this.
- Do not design blocking wait behavior into v1 unless the spec changes.
- Scan-launch tools should return clear normalized state immediately.

## Documentation Discipline

Keep docs aligned with implementation.

Update docs when:
- a tool contract changes,
- a safety classification changes,
- a config surface changes,
- a real ASoC endpoint differs from the original assumption,
- a spec item is deferred or promoted.

If reality diverges from `docs/appscan-asoc-mcp-server-spec.md`, do not silently improvise. Record the discrepancy and resolve it explicitly.

## Definition of Done

A task is not done until all of the following are true:
- Failing test written first.
- Implementation passes targeted tests.
- Full suite passes.
- Safety metadata added or updated.
- Docs updated if behavior changed.
- Code is formatted.
- Commit created if the change is a logical checkpoint.

## Recommended delivery sequence

1. Initialize repository layout and framework wiring.
2. Add config package with tests.
3. Add readonly package with tests.
4. Add ASoC client with auth/header tests and error mapping tests.
5. Add normalized models and mapper tests.
6. Implement applications tool group.
7. Implement files upload + template workflow.
8. Implement scans tool group.
9. Implement findings tool group.
10. Implement reports tool group.
11. Add optional policy/compliance reads.
12. Add optional admin tools with conservative safety profiles.
13. Add live smoke tests.
14. Tighten README and examples.

## Things to watch carefully

- ASoC endpoint behavior may differ subtly across tenants or license tiers.
- Findings may contain sensitive data; avoid unnecessary logging.
- Upload and scan-launch tools must be classified as high-risk/write tools.
- Do not let normalization erase useful vendor-specific details.
- Avoid overengineering for SAST/SCA/IAST now; keep extension points clean instead.
- Keep handlers small and focused; heavy logic belongs in client and normalize layers.
- Report binary retrieval should be added if straightforward, but not at the cost of destabilizing the core scan/findings/report flow.

## Preferred working style in OpenCode

When working in OpenCode:
- start each feature by naming the subtask,
- inspect the relevant spec section first,
- write the failing test,
- implement the smallest passing change,
- refactor,
- run package tests,
- run full tests before commit,
- commit at the checkpoint,
- then move to the next subtask.

## Escalate / ask when needed

Ask for clarification before proceeding if any of these are unclear:
- exact tenant base URL patterns,
- whether any tenant-specific API quirks need to override the spec,
- whether retries should be conservative or aggressive for ASoC calls,
- whether a supposedly straightforward binary report path turns out to be unexpectedly complex.

When in doubt, choose the simpler implementation that preserves correctness, testability, and safety reporting.
