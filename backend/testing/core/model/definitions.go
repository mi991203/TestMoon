package model

import (
	p "github.com/mi991203/TestMoon/libs/go/platform"
)

// Resource is the shared catalog identity. Each definition lives in its own
// concrete table; the association from a case to an API is indexed explicitly.
type DefinitionOwner struct {
	p.Base
	OrgID     string `gorm:"type:char(36);not null;index:definition_scope"`
	ProjectID string `gorm:"type:char(36);not null;index:definition_scope"`
}
type APIDefinition struct {
	DefinitionOwner
	EnvironmentID string `gorm:"type:char(36);not null;default:''"`
	RunnerID      string `gorm:"size:36;not null;default:''"`
	Protocol      string `gorm:"size:24;not null"`
	Description   string `gorm:"type:text"`
	RequestCipher string `gorm:"type:longtext;not null"`
	SchemaCipher  string `gorm:"type:longtext"`
}
type TestCaseDefinition struct {
	DefinitionOwner
	EnvironmentID             string `gorm:"type:char(36);not null;default:''"`
	RunnerID                  string `gorm:"size:36;not null;default:''"`
	Name                      string `gorm:"size:128;not null;default:''"`
	APIID                     string `gorm:"type:char(36);index:case_api;not null;default:''"`
	RequestCipher             string `gorm:"type:longtext"`
	OverrideCipher            string `gorm:"type:longtext"`
	VariablesCipher           string `gorm:"type:longtext"`
	CollectionVariablesCipher string `gorm:"type:longtext"`
	BeforeCipher              string `gorm:"type:longtext"`
	AfterCipher               string `gorm:"type:longtext"`
	AssertionsCipher          string `gorm:"type:longtext"`
	ExtractionsCipher         string `gorm:"type:longtext"`
}
type ScenarioDefinition struct {
	DefinitionOwner
	Name                      string `gorm:"size:128;not null;default:''"`
	StepsCipher               string `gorm:"type:longtext;not null"`
	CollectionVariablesCipher string `gorm:"type:longtext"`
}
type EnvironmentDefinition struct {
	DefinitionOwner
	VariablesCipher string `gorm:"type:longtext;not null"`
}
type ResourceFolder struct {
	DefinitionOwner
	AssetKind string `gorm:"size:24;not null"`
}
type SchemaMigration struct {
	Version string `gorm:"size:128;primaryKey"`
}
