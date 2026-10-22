package api

import (
	"context"
	"encoding/json"
	"time"

	"github.com/mi991203/TestMoon/backend/testing/execution/dao"
	"github.com/mi991203/TestMoon/backend/testing/execution/logic"
	"github.com/mi991203/TestMoon/backend/testing/execution/model"
	"github.com/mi991203/TestMoon/backend/testing/execution/serializer"
	p "github.com/mi991203/TestMoon/libs/go/platform"
	m "github.com/mi991203/TestMoon/libs/go/testing/domain"
	pb "github.com/mi991203/TestMoon/protocols/testing/execution/v1"
)

func (s *Server) ListRuns(ctx context.Context, r *pb.ListRunsRequest) (*pb.ListRunsResponse, error) {
	if err := validateRequest(r); err != nil {
		return nil, err
	}
	deps := s.deps
	if e := deps.Core.Authorize(ctx, r.OrgId, r.ProjectId, "read"); e != nil {
		return nil, e
	}
	var items []model.Run
	if e := dao.WithProjectScope(deps.DB, r.OrgId, r.ProjectId).Order("created_at DESC").Limit(100).Find(&items).Error; e != nil {
		return nil, p.DBError(e)
	}
	return serializer.ListRuns(map[string]any{"items": items})
}

func (s *Server) GetRun(ctx context.Context, r *pb.GetRunRequest) (*pb.GetRunResponse, error) {
	if err := validateRequest(r); err != nil {
		return nil, err
	}
	deps := s.deps
	if e := deps.Core.Authorize(ctx, r.OrgId, r.ProjectId, "read"); e != nil {
		return nil, e
	}
	var run model.Run
	if e := dao.WithProjectScope(deps.DB, r.OrgId, r.ProjectId).First(&run, "id = ?", r.Id).Error; e != nil {
		return nil, p.DBError(e)
	}
	var result any
	if run.Result != "" {
		raw, e := p.Open(deps.Config.EncryptionKey, run.Result)
		if e != nil {
			return nil, e
		}
		_ = json.Unmarshal([]byte(raw), &result)
	}
	return serializer.GetRun(map[string]any{"run": run, "result": result})
}

func (s *Server) DebugHTTP(ctx context.Context, r *pb.DebugHTTPRequest) (*pb.DebugHTTPResponse, error) {
	if err := validateRequest(r); err != nil {
		return nil, err
	}
	deps := s.deps
	if err := deps.Core.Authorize(ctx, r.OrgId, r.ProjectId, "execute"); err != nil {
		return nil, err
	}
	if r.Debug == nil {
		return nil, p.Bad("缺少调试请求")
	}
	input := &logic.RunInput{}
	if err := p.Decode(r, input); err != nil {
		return nil, err
	}

	out, err := logic.StartRun(ctx, deps, input)
	if err != nil {
		return nil, err
	}
	return &pb.DebugHTTPResponse{Run: serializer.Run(out)}, nil
}

func (s *Server) RunCase(ctx context.Context, r *pb.RunCaseRequest) (*pb.RunCaseResponse, error) {
	if err := validateRequest(r); err != nil {
		return nil, err
	}
	deps := s.deps
	if err := deps.Core.Authorize(ctx, r.OrgId, r.ProjectId, "execute"); err != nil {
		return nil, err
	}
	if r.CaseId == "" {
		return nil, p.Bad("缺少执行目标")
	}
	input := &logic.RunInput{}
	if err := p.Decode(r, input); err != nil {
		return nil, err
	}
	input.Items = []m.Selection{{Kind: "case", ID: r.CaseId}}
	out, err := logic.StartRun(ctx, deps, input)
	if err != nil {
		return nil, err
	}
	return &pb.RunCaseResponse{Run: serializer.Run(out)}, nil
}

func (s *Server) RunScenario(ctx context.Context, r *pb.RunScenarioRequest) (*pb.RunScenarioResponse, error) {
	if err := validateRequest(r); err != nil {
		return nil, err
	}
	deps := s.deps
	if err := deps.Core.Authorize(ctx, r.OrgId, r.ProjectId, "execute"); err != nil {
		return nil, err
	}
	if r.ScenarioId == "" {
		return nil, p.Bad("缺少执行目标")
	}
	input := &logic.RunInput{}
	if err := p.Decode(r, input); err != nil {
		return nil, err
	}
	input.Items = []m.Selection{{Kind: "scenario", ID: r.ScenarioId}}
	out, err := logic.StartRun(ctx, deps, input)
	if err != nil {
		return nil, err
	}
	return &pb.RunScenarioResponse{Run: serializer.Run(out)}, nil
}

func (s *Server) RunTask(ctx context.Context, r *pb.RunTaskRequest) (*pb.RunTaskResponse, error) {
	if err := validateRequest(r); err != nil {
		return nil, err
	}
	deps := s.deps
	if err := deps.Core.Authorize(ctx, r.OrgId, r.ProjectId, "execute"); err != nil {
		return nil, err
	}
	if r.TaskId == "" {
		return nil, p.Bad("缺少执行目标")
	}
	input := &logic.RunInput{}
	if err := p.Decode(r, input); err != nil {
		return nil, err
	}
	input.Id = r.TaskId
	out, err := logic.StartRun(ctx, deps, input)
	if err != nil {
		return nil, err
	}
	return &pb.RunTaskResponse{Run: serializer.Run(out)}, nil
}

func (s *Server) CancelRun(ctx context.Context, r *pb.CancelRunRequest) (*pb.CancelRunResponse, error) {
	if err := validateRequest(r); err != nil {
		return nil, err
	}
	deps := s.deps
	if e := deps.Core.Authorize(ctx, r.OrgId, r.ProjectId, "execute"); e != nil {
		return nil, e
	}
	now := time.Now()
	result := dao.WithProjectScope(deps.DB.Model(&model.Run{}), r.OrgId, r.ProjectId).Where("id = ? AND state IN ?", r.Id, []string{"created", "queued", "running"}).Updates(map[string]any{"state": "canceled", "finished_at": now, "error": "用户取消"})
	if result.Error != nil {
		return nil, p.DBError(result.Error)
	}
	return serializer.CancelRun(map[string]any{"ok": true})
}
