// Package runtime owns service infrastructure. Business functions receive these
// dependencies explicitly; API servers only own transport orchestration.
package runtime

import (
	"gorm.io/gorm"

	"github.com/mi991203/TestMoon/backend/testing/mock/conf"
	"github.com/mi991203/TestMoon/backend/testing/mock/dao"
	"github.com/mi991203/TestMoon/backend/testing/mock/thirdparty"
	p "github.com/mi991203/TestMoon/libs/go/platform"
)

type Dependencies struct {
	Config p.Config
	DB     *gorm.DB
	Core   *thirdparty.CoreClient
	Rules  *dao.Rules
}

func New(c conf.Config) *Dependencies {
	deps := &Dependencies{Config: c, DB: dao.Open(c)}
	deps.Core = thirdparty.NewCoreClient(c)
	deps.Rules = dao.NewRules(c.EncryptionKey)
	if err := deps.Rules.Migrate(deps.DB); err != nil {
		panic(err)
	}
	return deps
}
