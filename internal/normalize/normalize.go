package normalize

import (
	"github.com/karldane/appscan-asoc-mcp/internal/model"
	"time"
)

func Application(raw map[string]any) *model.Application {
	app := &model.Application{
		Raw: raw,
	}

	if v, ok := raw["Id"].(string); ok {
		app.ID = v
	}
	if v, ok := raw["Name"].(string); ok {
		app.Name = v
	}
	if v, ok := raw["Description"].(string); ok {
		app.Description = &v
	}
	if v, ok := raw["BusinessUnit"].(string); ok {
		app.BusinessUnit = &v
	}
	if v, ok := raw["Tags"].([]any); ok {
		app.Tags = sliceToStrings(v)
	}
	if v, ok := raw["AssetGroupIds"].([]any); ok {
		app.AssetGroupIDs = sliceToStrings(v)
	}
	if v, ok := raw["Created"].(string); ok {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			app.CreatedAt = &t
		}
	}
	if v, ok := raw["LastModified"].(string); ok {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			app.UpdatedAt = &t
		}
	}

	return app
}

func Scan(raw map[string]any) *model.Scan {
	scan := &model.Scan{
		ScanFamily: "dast",
		Status:     "unknown",
		QueueState: "unknown",
		Raw:        raw,
	}

	if v, ok := raw["Id"].(string); ok {
		scan.ID = v
	}
	if v, ok := raw["ApplicationId"].(string); ok {
		scan.AppID = &v
	}
	if v, ok := raw["ScanType"].(string); ok {
		scan.ScanFamily = v
	}
	if v, ok := raw["State"].(string); ok {
		scan.Status = normalizeStatus(v)
	}
	if v, ok := raw["ExecutionStatus"].(string); ok {
		scan.QueueState = normalizeQueueState(v)
	}
	if v, ok := raw["Url"].(string); ok {
		scan.Target = &v
	}
	if v, ok := raw["SubmissionTime"].(string); ok {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			scan.SubmittedAt = &t
		}
	}
	if v, ok := raw["StartTime"].(string); ok {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			scan.StartedAt = &t
		}
	}
	if v, ok := raw["FinishTime"].(string); ok {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			scan.CompletedAt = &t
		}
	}

	return scan
}

func Finding(raw map[string]any) *model.Finding {
	f := &model.Finding{
		ScanFamily: "dast",
		Status:     "unknown",
		Severity:   "unknown",
		Raw:        raw,
	}

	if v, ok := raw["Id"].(string); ok {
		f.ID = v
	}
	if v, ok := raw["ApplicationId"].(string); ok {
		f.AppID = &v
	}
	if v, ok := raw["ScanId"].(string); ok {
		f.ScanID = &v
	}
	if v, ok := raw["IssueName"].(string); ok {
		f.Title = v
	}
	if v, ok := raw["Severity"].(string); ok {
		f.Severity = normalizeSeverity(v)
	}
	if v, ok := raw["Status"].(string); ok {
		f.Status = normalizeFindingStatus(v)
	}
	if v, ok := raw["IssueType"].(string); ok {
		f.IssueType = &v
	}
	if v, ok := raw["Location"].(string); ok {
		f.Location = &v
	}
	if v, ok := raw["FindingStatus"].(string); ok {
		f.Status = normalizeFindingStatus(v)
	}
	if v, ok := raw["VulnerabilityName"].(string); ok {
		f.Title = v
	}
	if v, ok := raw["Url"].(string); ok {
		f.Location = &v
	}

	return f
}

func Report(raw map[string]any) *model.Report {
	r := &model.Report{
		Status: "unknown",
		Format: "unknown",
		Raw:    raw,
	}

	if v, ok := raw["Id"].(string); ok {
		r.ID = v
	}
	if v, ok := raw["ApplicationId"].(string); ok {
		r.AppID = &v
	}
	if v, ok := raw["ScanId"].(string); ok {
		r.ScanID = &v
	}
	if v, ok := raw["Status"].(string); ok {
		r.Status = normalizeReportStatus(v)
	}
	if v, ok := raw["ReportType"].(string); ok {
		r.Format = v
	}
	if v, ok := raw["DownloadUrl"].(string); ok {
		r.DownloadURL = &v
	}
	if v, ok := raw["CreatedDate"].(string); ok {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			r.CreatedAt = &t
		}
	}

	return r
}

func normalizeStatus(s string) string {
	switch s {
	case "Ready", "Completed":
		return "completed"
	case "Running":
		return "running"
	case "Failed":
		return "failed"
	case "Queued":
		return "queued"
	case "Stopped", "Canceled":
		return "canceled"
	default:
		return "unknown"
	}
}

func normalizeQueueState(s string) string {
	switch s {
	case "Queued":
		return "queued"
	case "Started", "Running", "Ready":
		return "not_queued"
	default:
		return "unknown"
	}
}

func normalizeSeverity(s string) string {
	switch s {
	case "Critical", "High":
		return "high"
	case "Medium":
		return "medium"
	case "Low":
		return "low"
	case "Info", "Informational":
		return "info"
	default:
		return "unknown"
	}
}

func normalizeFindingStatus(s string) string {
	switch s {
	case "Open", "New", "Active":
		return "open"
	case "Fixed", "Closed":
		return "fixed"
	case "Ignored", "False Positive":
		return "ignored"
	case "Noise":
		return "noise"
	default:
		return "unknown"
	}
}

func normalizeReportStatus(s string) string {
	switch s {
	case "Ready", "Completed":
		return "ready"
	case "In Progress", "Running", "Pending":
		return "pending"
	case "Failed", "Error":
		return "failed"
	default:
		return "unknown"
	}
}

func sliceToStrings(slice []any) []string {
	result := make([]string, 0, len(slice))
	for _, v := range slice {
		if s, ok := v.(string); ok {
			result = append(result, s)
		}
	}
	return result
}
