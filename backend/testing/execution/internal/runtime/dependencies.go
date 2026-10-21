// Package runtime owns service infrastructure. Business functions receive these
// dependencies explicitly; API servers only own transport orchestration.
package runtime

import (
	"github.com/twmb/franz-go/pkg/kgo"
	"gorm.io/gorm"

	"github.com/mi991203/TestMoon/backend/testing/execution/conf"
	"github.com/mi991203/TestMoon/backend/testing/execution/dao"
	"github.com/mi991203/TestMoon/backend/testing/execution/thirdparty"
	p "github.com/mi991203/TestMoon/libs/go/platform"
)

type Dependencies struct {
	Config   p.Config
	DB       *gorm.DB
	Core     *thirdparty.CoreClient
	Tasks    *dao.Tasks
	Producer *kgo.Client
	Consumer *kgo.Client
}

func New(c conf.Config) *Dependencies {
	deps := &Dependencies{Config: c, DB: dao.Open(c)}
	deps.Core = thirdparty.NewCoreClient(c)
	deps.Tasks = dao.NewTasks(c.EncryptionKey)
	if err := deps.Tasks.Migrate(deps.DB); err != nil {
		panic(err)
	}
	return deps
}
