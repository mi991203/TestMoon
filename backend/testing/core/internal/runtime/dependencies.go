// Package runtime owns service infrastructure. Business functions receive these
// dependencies explicitly; API servers only own transport orchestration.
package runtime

import (
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"github.com/mi991203/TestMoon/backend/testing/core/conf"
	"github.com/mi991203/TestMoon/backend/testing/core/dao"
	"github.com/mi991203/TestMoon/backend/testing/core/thirdparty"
	p "github.com/mi991203/TestMoon/libs/go/platform"
)

type Dependencies struct {
	Config p.Config
	DB     *gorm.DB
	Redis  *redis.Client
	Assets *dao.Resources
	Remote *thirdparty.Clients
}

func New(c conf.Config) *Dependencies {
	deps := &Dependencies{Config: c, DB: dao.Open(c)}
	deps.Redis = p.Redis(c)
	deps.Assets = dao.NewResources(c.EncryptionKey)
	if err := deps.Assets.Migrate(deps.DB); err != nil {
		panic(err)
	}
	deps.Remote = thirdparty.New(c)
	return deps
}
