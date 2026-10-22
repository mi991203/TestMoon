package api

import (
	"context"

	"github.com/mi991203/TestMoon/backend/testing/core/logic"
	"github.com/mi991203/TestMoon/backend/testing/core/serializer"
	p "github.com/mi991203/TestMoon/libs/go/platform"
	pb "github.com/mi991203/TestMoon/protocols/testing/core/v1"
)

func (s *Server) ListScenarios(ctx context.Context, r *pb.ListScenariosRequest) (*pb.ListScenariosResponse, error) {
	if err := validateRequest(r); err != nil {
		return nil, err
	}
	deps := s.deps
	items, err := logic.ListResources(ctx, deps, r.GetOrgId(), r.GetProjectId(), "scenario", r.Search, r.FolderId, nil, "")
	if err != nil {
		return nil, err
	}
	return serializer.ListScenarios(map[string]any{"items": items})
}

func (s *Server) GetScenario(ctx context.Context, r *pb.GetScenarioRequest) (*pb.GetScenarioResponse, error) {
	if err := validateRequest(r); err != nil {
		return nil, err
	}
	deps := s.deps
	items, err := logic.ListResources(ctx, deps, r.GetOrgId(), r.GetProjectId(), "scenario", "", nil, nil, r.Id)
	if err != nil {
		return nil, err
	}
	if len(items) != 1 {
		return nil, p.NotFound()
	}
	return serializer.GetScenario(map[string]any{"resource": items[0]})
}

func (s *Server) CreateScenario(ctx context.Context, r *pb.CreateScenarioRequest) (*pb.CreateScenarioResponse, error) {
	if err := validateRequest(r); err != nil {
		return nil, err
	}
	deps := s.deps
	in, err := resourceWriteInput(r, "scenario")
	if err != nil {
		return nil, err
	}
	item, err := logic.SaveResource(ctx, deps, in)
	if err != nil {
		return nil, err
	}
	return serializer.CreateScenario(map[string]any{"resource": item})
}

func (s *Server) UpdateScenario(ctx context.Context, r *pb.UpdateScenarioRequest) (*pb.UpdateScenarioResponse, error) {
	if err := validateRequest(r); err != nil {
		return nil, err
	}
	deps := s.deps
	in, err := resourceWriteInput(r, "scenario")
	if err != nil {
		return nil, err
	}
	item, err := logic.SaveResource(ctx, deps, in)
	if err != nil {
		return nil, err
	}
	return serializer.UpdateScenario(map[string]any{"resource": item})
}

func (s *Server) DeleteScenario(ctx context.Context, r *pb.DeleteScenarioRequest) (*pb.DeleteScenarioResponse, error) {
	if err := validateRequest(r); err != nil {
		return nil, err
	}
	deps := s.deps
	in := &logic.ResourceInput{OrgId: r.OrgId, ProjectId: r.ProjectId, Id: r.Id, Kind: "scenario"}
	if err := logic.DeleteResource(ctx, deps, in); err != nil {
		return nil, err
	}
	return serializer.DeleteScenario(map[string]any{"ok": true})
}
