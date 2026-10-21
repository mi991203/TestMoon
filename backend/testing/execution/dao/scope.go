package dao

import (
	"gorm.io/gorm"
)

// WithProjectScope limits a query to the supplied organization and project.
func WithProjectScope(db *gorm.DB, org, project string) *gorm.DB {
	return db.Where("org_id = ? AND project_id = ?", org, project)
}
