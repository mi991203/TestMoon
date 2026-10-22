package dao

import (
	"gorm.io/gorm"

	"github.com/mi991203/TestMoon/backend/testing/core/conf"
	"github.com/mi991203/TestMoon/backend/testing/core/model"
	p "github.com/mi991203/TestMoon/libs/go/platform"
)

// Open owns connection setup and schema registration for this service.
func Open(c conf.Config) *gorm.DB {
	return p.DB(c, &model.Project{}, &model.Member{}, &model.Resource{}, &model.Share{}, &model.ImportBatch{})
}
