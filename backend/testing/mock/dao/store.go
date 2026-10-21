package dao

import (
	"gorm.io/gorm"

	"github.com/mi991203/TestMoon/backend/testing/mock/conf"
	"github.com/mi991203/TestMoon/backend/testing/mock/model"
	p "github.com/mi991203/TestMoon/libs/go/platform"
)

// Open owns connection setup and schema registration for this service.
func Open(c conf.Config) *gorm.DB { return p.DB(c, &model.Endpoint{}, &model.Rule{}) }
