package model

import (
	"time"

	p "github.com/mi991203/TestMoon/libs/go/platform"
)

type Task struct {
	p.Base
	OrgID     string `gorm:"type:char(36);index" json:"org_id"`
	ProjectID string `gorm:"type:char(36);index" json:"project_id"`
	Name      string `json:"name"`
	Revision  int64  `json:"revision"`
	// Document is transient compatibility input; persisted fields are explicit below.
	Document       string `gorm:"-" json:"-"`
	LegacyDocument string `gorm:"column:document;type:longtext" json:"-"`
	ItemsJSON      string `gorm:"type:json" json:"-"`
	EnvironmentID  string `gorm:"type:char(36);index" json:"-"`
	RunnerID       string `gorm:"size:36;index" json:"-"`
	TimeoutSeconds int    `json:"-"`
	ConfigVersion  int    `gorm:"not null;default:0" json:"-"`
}
type Run struct {
	p.Base
	OrgID          string     `gorm:"type:char(36);index:run_scope" json:"org_id"`
	ProjectID      string     `gorm:"type:char(36);index:run_scope" json:"project_id"`
	UserID         string     `json:"user_id"`
	TaskID         string     `json:"task_id"`
	Name           string     `json:"name"`
	State          string     `gorm:"size:24;index" json:"state"`
	TargetRunnerID string     `gorm:"size:36;index" json:"target_runner_id"`
	NodeID         string     `gorm:"type:char(36);index" json:"node_id"`
	RequestHash    string     `gorm:"size:64;uniqueIndex" json:"-"`
	Fingerprint    string     `gorm:"size:64" json:"-"`
	LeaseHash      string     `gorm:"size:64" json:"-"`
	Snapshot       string     `gorm:"type:longtext" json:"-"`
	Result         string     `gorm:"type:longtext" json:"-"`
	TimeoutSeconds int        `json:"timeout_seconds"`
	QueueDeadline  time.Time  `json:"-"`
	LeaseUntil     *time.Time `json:"-"`
	Deadline       *time.Time `json:"-"`
	StartedAt      *time.Time `json:"started_at"`
	FinishedAt     *time.Time `json:"finished_at"`
	Error          string     `gorm:"size:256" json:"error,omitempty"`
}
type Runner struct {
	p.Base
	OrgID          string     `gorm:"type:char(36);index" json:"org_id"`
	ProjectID      string     `gorm:"type:char(36);index" json:"project_id"`
	Name           string     `json:"name"`
	Public         bool       `json:"public"`
	Revoked        bool       `json:"revoked"`
	EnrollmentHash string     `gorm:"size:64;uniqueIndex" json:"-"`
	CredentialHash string     `gorm:"size:64;index" json:"-"`
	Registered     bool       `json:"registered"`
	LastSeen       *time.Time `json:"last_seen"`
	Concurrency    int        `json:"concurrency"`
	Version        string     `json:"version"`
	Capabilities   string     `json:"-"`
}
type Budget struct {
	OrgID string `gorm:"type:char(36);primaryKey"`
}
type Outbox struct {
	p.Base
	RunID      string `gorm:"type:char(36);uniqueIndex"`
	Payload    string `gorm:"type:text"`
	Published  bool   `gorm:"index"`
	LeaseUntil time.Time
}
type ConsumedEvent struct {
	ID        string `gorm:"type:char(36);primaryKey"`
	CreatedAt time.Time
}
