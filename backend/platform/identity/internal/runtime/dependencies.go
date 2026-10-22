// Package runtime owns service infrastructure. Business functions receive these
// dependencies explicitly; API servers only own transport orchestration.
package runtime

import (
	"context"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"github.com/mi991203/TestMoon/backend/platform/identity/conf"
	"github.com/mi991203/TestMoon/backend/platform/identity/dao"
	p "github.com/mi991203/TestMoon/libs/go/platform"
)

type Dependencies struct {
	Config p.Config
	DB     *gorm.DB
	Redis  *redis.Client
}

func New(c conf.Config) *Dependencies {
	deps := &Dependencies{Config: c, DB: dao.Open(c)}
	deps.Redis = p.Redis(c)
	if err := deps.Redis.Ping(context.Background()).Err(); err != nil {
		panic("identity Redis unavailable")
	}
	return deps
}
