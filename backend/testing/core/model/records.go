package model

import (
	"time"

	p "github.com/mi991203/TestMoon/libs/go/platform"
)

type Project struct {
	p.Base
	OrgID       string `gorm:"type:char(36);index" json:"org_id"`
	Name        string `gorm:"size:128" json:"name"`
	Description string `gorm:"type:text" json:"description"`
}
type Member struct {
	p.Base
	OrgID     string `gorm:"type:char(36);uniqueIndex:project_member" json:"org_id"`
	ProjectID string `gorm:"type:char(36);uniqueIndex:project_member" json:"project_id"`
	UserID    string `gorm:"type:char(36);uniqueIndex:project_member" json:"user_id"`
	Role      string `gorm:"size:16" json:"role"`
}
type Resource struct {
	p.Base
	OrgID     string `gorm:"type:char(36);index:resource_scope" json:"org_id"`
	ProjectID string `gorm:"type:char(36);index:resource_scope" json:"project_id"`
	Kind      string `gorm:"size:24;index:resource_scope" json:"kind"`
	FolderID  string `gorm:"type:char(36);index" json:"folder_id"`
	Name      string `gorm:"size:128" json:"name"`
	Revision  int64  `json:"revision"`
	Document  string `gorm:"-" json:"-"`
	// LegacyDocument is retained only for the idempotent compatibility migration.
	LegacyDocument string `gorm:"column:document;type:longtext" json:"-"`
}
type Share struct {
	p.Base
	OrgID     string     `gorm:"type:char(36);index" json:"org_id"`
	ProjectID string     `gorm:"type:char(36);index" json:"project_id"`
	Name      string     `json:"name"`
	TokenHash string     `gorm:"size:64;uniqueIndex" json:"-"`
	Token     string     `gorm:"type:text" json:"-"`
	Password  string     `gorm:"size:256" json:"-"`
	APIIDs    string     `gorm:"type:text" json:"-"`
	ExpiresAt *time.Time `json:"expires_at"`
}
type ImportBatch struct {
	p.Base
	OrgID     string `gorm:"type:char(36);index"`
	ProjectID string `gorm:"type:char(36);index"`
	UserID    string `gorm:"type:char(36)"`
	Plan      string `gorm:"type:longtext"`
	ExpiresAt time.Time
	Used      bool
}
type ImportEntry struct {
	TempID   string         `json:"temp_id"`
	Kind     string         `json:"kind"`
	Name     string         `json:"name"`
	Folder   string         `json:"folder"`
	Document map[string]any `json:"document"`
}
type ImportPlan struct {
	Entries     []ImportEntry `json:"entries"`
	Diagnostics []string      `json:"diagnostics"`
}
