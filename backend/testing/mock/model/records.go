package model

import (
	p "github.com/mi991203/TestMoon/libs/go/platform"
)

type Endpoint struct {
	p.Base
	OrgID     string `gorm:"type:char(36);index" json:"org_id"`
	ProjectID string `gorm:"type:char(36);index" json:"project_id"`
	Name      string `gorm:"size:128" json:"name"`
}
type Rule struct {
	p.Base
	OrgID             string `gorm:"type:char(36);index" json:"org_id"`
	ProjectID         string `gorm:"type:char(36);index" json:"project_id"`
	EndpointID        string `gorm:"type:char(36);index" json:"endpoint_id"`
	Name              string `gorm:"size:128" json:"name"`
	Method            string `gorm:"size:12" json:"method"`
	Path              string `gorm:"size:1024" json:"path"`
	Priority          int    `json:"priority"`
	Document          string `gorm:"-" json:"-"`
	LegacyDocument    string `gorm:"column:document;type:longtext" json:"-"`
	Status            int    `json:"-"`
	DelayMS           int    `json:"-"`
	HeadersCipher     string `gorm:"type:longtext" json:"-"`
	BodyCipher        string `gorm:"type:longtext" json:"-"`
	ConditionsCipher  string `gorm:"type:longtext" json:"-"`
	DefinitionVersion int    `gorm:"not null;default:0" json:"-"`
}
type Condition struct {
	Source   string `json:"source"`
	Path     string `json:"path"`
	Operator string `json:"operator"`
	Expected any    `json:"expected"`
}
type Response struct {
	Conditions []Condition       `json:"conditions"`
	Status     int               `json:"status"`
	Headers    map[string]string `json:"headers"`
	Body       string            `json:"body"`
	DelayMS    int               `json:"delay_ms"`
}
