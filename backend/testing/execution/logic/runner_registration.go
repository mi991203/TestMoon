package logic

import (
	"context"

	"github.com/mi991203/TestMoon/backend/testing/execution/internal/runtime"
	"github.com/mi991203/TestMoon/backend/testing/execution/model"
	p "github.com/mi991203/TestMoon/libs/go/platform"
)

func CreateRunner(ctx context.Context, deps *runtime.Dependencies, org, project, name string, concurrency int, public bool) (model.Runner, string, error) {
	if public {
		if e := p.RequireAdmin(ctx); e != nil {
			return model.Runner{}, "", e
		}
	}
	token := p.Token()
	v := model.Runner{Base: p.NewBase(), Name: name, OrgID: org, ProjectID: project, Public: public, Concurrency: concurrency, EnrollmentHash: p.Hash(token)}
	if public {
		v.OrgID = ""
		v.ProjectID = ""
	}
	if e := deps.DB.Create(&v).Error; e != nil {
		return model.Runner{}, "", p.DBError(e)
	}
	return v, token, nil
}
