package downloader

import "time"

type TaskKind string

const (
	KindVideo    TaskKind = "video"
	KindSubtitle TaskKind = "subtitle"
)

type Status string

const (
	StatusQueued      Status = "queued"
	StatusDownloading Status = "downloading"
	StatusCompleted   Status = "completed"
	StatusFailed      Status = "failed"
)

type Task struct {
	ID          string    `json:"id"`
	ParentID    string    `json:"parent_id,omitempty"`
	Kind        TaskKind  `json:"kind"`
	ItemID      string    `json:"item_id"`
	SourceID    string    `json:"source_id"`
	StreamIndex int       `json:"stream_index"`
	DisplayName string    `json:"display_name"`
	URL         string    `json:"url"`
	OutputPath  string    `json:"output_path"`
	TotalSize   int64     `json:"total_size"`
	Downloaded  int64     `json:"downloaded"`
	Status      Status    `json:"status"`
	Attempts    int       `json:"attempts"`
	LastError   string    `json:"last_error,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
