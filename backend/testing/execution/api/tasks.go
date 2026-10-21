package api

import (
	"context"
	"encoding/json"

	"github.com/mi991203/TestMoon/backend/testing/execution/dao"
	"github.com/mi991203/TestMoon/backend/testing/execution/logic"
	"github.com/mi991203/TestMoon/backend/testing/execution/model"
	"github.com/mi991203/TestMoon/backend/testing/execution/serializer"
	p "github.com/mi991203/TestMoon/libs/go/platform"
	m "github.com/mi991203/TestMoon/libs/go/testing/domain"
	core "github.com/mi991203/TestMoon/protocols/testing/core/v1"
	pb "github.com/mi991203/TestMoon/protocols/testing/execution/v1"
)

func (s *Server) ListTasks(ctx context.Context, r *pb.ListTasksRequest) (*pb.ListTasksResponse, error) {
	if err := validateRequest(r); err != nil {
		return nil, err
	}
	deps := s.deps
	if e := deps.Core.Authorize(ctx, r.OrgId, r.ProjectId, "read"); e != nil {
		return nil, e
	}
	var all []model.Task
	if e := dao.WithProjectScope(deps.DB, r.OrgId, r.ProjectId).Order("created_at DESC").Find(&all).Error; e != nil {
		return nil, p.DBError(e)
	}
	items := []any{}
	for _, t := range all {
		doc, e := deps.Tasks.Definition(t)
		if e != nil {
			return nil, e
		}
		items = append(items, map[string]any{"id": t.ID, "name": t.Name, "revision": t.Revision, "document": doc})
	}
	return serializer.ListTasks(map[string]any{"items": items})
}

func (s *Server) CreateTask(ctx context.Context, r *pb.CreateTaskRequest) (*pb.CreateTaskResponse, error) {
	if err := validateRequest(r); err != nil {
		return nil, err
	}
	if r.Document == nil {
		return nil, p.Bad("缺少任务配置")
	}
	deps := s.deps
	if e := deps.Core.Authorize(ctx, r.OrgId, r.ProjectId, "write"); e != nil {
		return nil, e
	}
	var in struct {
		Name     string `json:"name" validate:"required,max=128"`
		Revision int64  `json:"revision"`
		Document m.Task `json:"document"`
	}
	if e := p.Decode(r, &in); e != nil {
		return nil, e
	}
	if _, e := deps.Core.Client.CreateSnapshot(p.Outgoing(ctx, deps.Config), &core.CreateSnapshotRequest{OrgId: r.OrgId, ProjectId: r.ProjectId, Items: r.GetDocument().GetItems(), EnvironmentId: r.GetDocument().GetEnvironmentId(), RunnerId: r.GetDocument().GetRunnerId(), TimeoutSeconds: r.GetDocument().GetTimeoutSeconds()}); e != nil {
		return nil, e
	}
	raw, _ := json.Marshal(in.Document)
	v := model.Task{Base: p.NewBase(), OrgID: r.OrgId, ProjectID: r.ProjectId, Name: in.Name, Revision: 1, Document: p.Seal(deps.Config.EncryptionKey, string(raw))}

	if e := deps.Tasks.Create(deps.DB, &v); e != nil {
		return nil, p.DBError(e)
	}

	return serializer.CreateTask(map[string]any{"id": v.ID})
}

func (s *Server) UpdateTask(ctx context.Context, r *pb.UpdateTaskRequest) (*pb.UpdateTaskResponse, error) {
	if err := validateRequest(r); err != nil {
		return nil, err
	}
	if r.Document == nil {
		return nil, p.Bad("缺少任务配置")
	}
	deps := s.deps
	if e := deps.Core.Authorize(ctx, r.OrgId, r.ProjectId, "write"); e != nil {
		return nil, e
	}
	var in struct {
		Name     string `json:"name" validate:"required,max=128"`
		Revision int64  `json:"revision"`
		Document m.Task `json:"document"`
	}
	if e := p.Decode(r, &in); e != nil {
		return nil, e
	}
	if _, e := deps.Core.Client.CreateSnapshot(p.Outgoing(ctx, deps.Config), &core.CreateSnapshotRequest{OrgId: r.OrgId, ProjectId: r.ProjectId, Items: r.GetDocument().GetItems(), EnvironmentId: r.GetDocument().GetEnvironmentId(), RunnerId: r.GetDocument().GetRunnerId(), TimeoutSeconds: r.GetDocument().GetTimeoutSeconds()}); e != nil {
		return nil, e
	}
	raw, _ := json.Marshal(in.Document)
	v := model.Task{Base: p.NewBase(), OrgID: r.OrgId, ProjectID: r.ProjectId, Name: in.Name, Revision: 1, Document: p.Seal(deps.Config.EncryptionKey, string(raw))}

	v.ID = r.Id
	if err := deps.Tasks.Update(deps.DB, v, in.Revision); err != nil {
		return nil, logic.ExecutionError(err)
	}

	return serializer.UpdateTask(map[string]any{"id": v.ID})
}

func (s *Server) DeleteTask(ctx context.Context, r *pb.DeleteTaskRequest) (*pb.DeleteTaskResponse, error) {
	if err := validateRequest(r); err != nil {
		return nil, err
	}
	deps := s.deps
	if e := deps.Core.Authorize(ctx, r.OrgId, r.ProjectId, "write"); e != nil {
		return nil, e
	}
	result := dao.WithProjectScope(deps.DB, r.OrgId, r.ProjectId).Where("id = ?", r.Id).Delete(&model.Task{})
	if result.Error != nil {
		return nil, p.DBError(result.Error)
	}
	return serializer.DeleteTask(map[string]any{"ok": true})
}
