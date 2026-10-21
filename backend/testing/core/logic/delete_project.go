package logic

import (
	"context"

	"gorm.io/gorm"

	"github.com/mi991203/TestMoon/backend/testing/core/dao"
	"github.com/mi991203/TestMoon/backend/testing/core/internal/runtime"
	"github.com/mi991203/TestMoon/backend/testing/core/model"
	p "github.com/mi991203/TestMoon/libs/go/platform"
	exec "github.com/mi991203/TestMoon/protocols/testing/execution/v1"
	mock "github.com/mi991203/TestMoon/protocols/testing/mock/v1"
)

func DeleteEmptyProject(ctx context.Context, deps *runtime.Dependencies, org, project string) error {
	execution := deps.Remote.Execution
	ctx = p.Outgoing(ctx, deps.Config)
	tasks, e := execution.ListTasks(ctx, &exec.ListTasksRequest{OrgId: org, ProjectId: project})
	if e != nil {
		return e
	}
	if len(tasks.Items) > 0 {
		return p.Conflict("项目仍有任务")
	}
	runners, e := execution.ListProjectRunners(ctx, &exec.ListProjectRunnersRequest{OrgId: org, ProjectId: project})
	if e != nil {
		return e
	}
	for _, runner := range runners.Items {
		if !runner.Public && !runner.Revoked {
			return p.Conflict("项目仍有有效Runner")
		}
	}
	runs, e := execution.ListRuns(ctx, &exec.ListRunsRequest{OrgId: org, ProjectId: project})
	if e != nil {
		return e
	}
	if len(runs.Items) > 0 {
		return p.Conflict("项目仍有执行记录")
	}
	endpoints, e := deps.Remote.Mock.ListEndpoints(ctx, &mock.ListEndpointsRequest{OrgId: org, ProjectId: project})
	if e != nil {
		return e
	}
	if len(endpoints.Items) > 0 {
		return p.Conflict("项目仍有Mock端点")
	}
	e = deps.DB.Transaction(func(tx *gorm.DB) error {
		if e := dao.WithProjectScope(tx, org, project).Delete(&model.Member{}).Error; e != nil {
			return e
		}
		if e := dao.WithProjectScope(tx, org, project).Delete(&model.Share{}).Error; e != nil {
			return e
		}
		return tx.Where("org_id = ? AND id = ?", org, project).Delete(&model.Project{}).Error
	})
	return p.DBError(e)
}
