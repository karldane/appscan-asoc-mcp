package model

import "time"

type Application struct {
	ID            string     `json:"id"`
	Name          string     `json:"name"`
	Description   *string    `json:"description"`
	BusinessUnit  *string    `json:"business_unit"`
	Tags          []string   `json:"tags"`
	AssetGroupIDs []string   `json:"asset_group_ids"`
	CreatedAt     *time.Time `json:"created_at"`
	UpdatedAt     *time.Time `json:"updated_at"`
	Raw           any        `json:"raw"`
}

type Scan struct {
	ID          string     `json:"id"`
	AppID       *string    `json:"app_id"`
	ScanFamily  string     `json:"scan_family"`
	Status      string     `json:"status"`
	QueueState  string     `json:"queue_state"`
	Target      *string    `json:"target"`
	SubmittedAt *time.Time `json:"submitted_at"`
	StartedAt   *time.Time `json:"started_at"`
	CompletedAt *time.Time `json:"completed_at"`
	Raw         any        `json:"raw"`
}

type Finding struct {
	ID             string     `json:"id"`
	AppID          *string    `json:"app_id"`
	ScanID         *string    `json:"scan_id"`
	ScanFamily     string     `json:"scan_family"`
	Title          string     `json:"title"`
	Severity       string     `json:"severity"`
	Status         string     `json:"status"`
	IssueType      *string    `json:"issue_type"`
	Location       *string    `json:"location"`
	FirstSeen      *time.Time `json:"first_seen"`
	LastSeen       *time.Time `json:"last_seen"`
	Compliance     []string   `json:"compliance"`
	Recommendation *string    `json:"recommendation"`
	Raw            any        `json:"raw"`
}

type Report struct {
	ID          string     `json:"id"`
	AppID       *string    `json:"app_id"`
	ScanID      *string    `json:"scan_id"`
	Status      string     `json:"status"`
	Format      string     `json:"format"`
	DownloadURL *string    `json:"download_url"`
	CreatedAt   *time.Time `json:"created_at"`
	Raw         any        `json:"raw"`
}

type FileInfo struct {
	ID        string     `json:"id"`
	Name      string     `json:"name"`
	Size      int64      `json:"size"`
	CreatedAt *time.Time `json:"created_at"`
	Raw       any        `json:"raw"`
}

type AssetGroup struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Raw  any    `json:"raw"`
}

type Policy struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description *string `json:"description"`
	Raw         any     `json:"raw"`
}
