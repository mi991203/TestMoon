package api

import (
	"context"

	"github.com/mi991203/TestMoon/backend/testing/core/logic"
	"github.com/mi991203/TestMoon/backend/testing/core/serializer"
	p "github.com/mi991203/TestMoon/libs/go/platform"
	pb "github.com/mi991203/TestMoon/protocols/testing/core/v1"
)

func (s *Server) ListAPIs(ctx context.Context, r *pb.ListAPIsRequest) (*pb.ListAPIsResponse, error) {
	if err := validateRequest(r); err != nil {
		return nil, err
	}
	deps := s.deps
	items, err := logic.ListResources(ctx, deps, r.GetOrgId(), r.GetProjectId(), "api", r.Search, r.FolderId, nil, "")
	if err != nil {
		return nil, err
	}
	return serializer.ListAPIs(map[string]any{"items": items})
}

func (s *Server) GetAPI(ctx context.Context, r *pb.GetAPIRequest) (*pb.GetAPIResponse, error) {
	if err := validateRequest(r); err != nil {
		return nil, err
	}
	deps := s.deps
	items, err := logic.ListResources(ctx, deps, r.GetOrgId(), r.GetProjectId(), "api", "", nil, nil, r.Id)
	if err != nil {
		return nil, err
	}
	if len(items) != 1 {
		return nil, p.NotFound()
	}
	return serializer.GetAPI(map[string]any{"resource": items[0]})
}

func (s *Server) CreateAPI(ctx context.Context, r *pb.CreateAPIRequest) (*pb.CreateAPIResponse, error) {
	if err := validateRequest(r); err != nil {
		return nil, err
	}
	deps := s.deps
	in, err := resourceWriteInput(r, "api")
	if err != nil {
		return nil, err
	}
	item, err := logic.SaveResource(ctx, deps, in)
	if err != nil {
		return nil, err
	}
	return serializer.CreateAPI(map[string]any{"resource": item})
}

func (s *Server) UpdateAPI(ctx context.Context, r *pb.UpdateAPIRequest) (*pb.UpdateAPIResponse, error) {
	if err := validateRequest(r); err != nil {
		return nil, err
	}
	deps := s.deps
	in, err := resourceWriteInput(r, "api")
	if err != nil {
		return nil, err
	}
	item, err := logic.SaveResource(ctx, deps, in)
	if err != nil {
		return nil, err
	}
	return serializer.UpdateAPI(map[string]any{"resource": item})
}

func (s *Server) DeleteAPI(ctx context.Context, r *pb.DeleteAPIRequest) (*pb.DeleteAPIResponse, error) {
	if err := validateRequest(r); err != nil {
		return nil, err
	}
	deps := s.deps
	in := &logic.ResourceInput{OrgId: r.OrgId, ProjectId: r.ProjectId, Id: r.Id, Kind: "api"}
	if err := logic.DeleteResource(ctx, deps, in); err != nil {
		return nil, err
	}
	return serializer.DeleteAPI(map[string]any{"ok": true})
}
