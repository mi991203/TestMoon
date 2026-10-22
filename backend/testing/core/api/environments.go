package api

import (
	"context"

	"github.com/mi991203/TestMoon/backend/testing/core/logic"
	"github.com/mi991203/TestMoon/backend/testing/core/serializer"
	p "github.com/mi991203/TestMoon/libs/go/platform"
	pb "github.com/mi991203/TestMoon/protocols/testing/core/v1"
)

func (s *Server) ListEnvironments(ctx context.Context, r *pb.ListEnvironmentsRequest) (*pb.ListEnvironmentsResponse, error) {
	if err := validateRequest(r); err != nil {
		return nil, err
	}
	deps := s.deps
	items, err := logic.ListResources(ctx, deps, r.GetOrgId(), r.GetProjectId(), "environment", r.Search, r.FolderId, nil, "")
	if err != nil {
		return nil, err
	}
	return serializer.ListEnvironments(map[string]any{"items": items})
}

func (s *Server) GetEnvironment(ctx context.Context, r *pb.GetEnvironmentRequest) (*pb.GetEnvironmentResponse, error) {
	if err := validateRequest(r); err != nil {
		return nil, err
	}
	deps := s.deps
	items, err := logic.ListResources(ctx, deps, r.GetOrgId(), r.GetProjectId(), "environment", "", nil, nil, r.Id)
	if err != nil {
		return nil, err
	}
	if len(items) != 1 {
		return nil, p.NotFound()
	}
	return serializer.GetEnvironment(map[string]any{"resource": items[0]})
}

func (s *Server) CreateEnvironment(ctx context.Context, r *pb.CreateEnvironmentRequest) (*pb.CreateEnvironmentResponse, error) {
	if err := validateRequest(r); err != nil {
		return nil, err
	}
	deps := s.deps
	in, err := resourceWriteInput(r, "environment")
	if err != nil {
		return nil, err
	}
	item, err := logic.SaveResource(ctx, deps, in)
	if err != nil {
		return nil, err
	}
	return serializer.CreateEnvironment(map[string]any{"resource": item})
}

func (s *Server) UpdateEnvironment(ctx context.Context, r *pb.UpdateEnvironmentRequest) (*pb.UpdateEnvironmentResponse, error) {
	if err := validateRequest(r); err != nil {
		return nil, err
	}
	deps := s.deps
	in, err := resourceWriteInput(r, "environment")
	if err != nil {
		return nil, err
	}
	item, err := logic.SaveResource(ctx, deps, in)
	if err != nil {
		return nil, err
	}
	return serializer.UpdateEnvironment(map[string]any{"resource": item})
}

func (s *Server) DeleteEnvironment(ctx context.Context, r *pb.DeleteEnvironmentRequest) (*pb.DeleteEnvironmentResponse, error) {
	if err := validateRequest(r); err != nil {
		return nil, err
	}
	deps := s.deps
	in := &logic.ResourceInput{OrgId: r.OrgId, ProjectId: r.ProjectId, Id: r.Id, Kind: "environment"}
	if err := logic.DeleteResource(ctx, deps, in); err != nil {
		return nil, err
	}
	return serializer.DeleteEnvironment(map[string]any{"ok": true})
}
