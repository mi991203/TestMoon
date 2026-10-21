package api

import (
	"context"

	"github.com/mi991203/TestMoon/backend/testing/core/logic"
	"github.com/mi991203/TestMoon/backend/testing/core/serializer"
	p "github.com/mi991203/TestMoon/libs/go/platform"
	pb "github.com/mi991203/TestMoon/protocols/testing/core/v1"
)

func (s *Server) ListTestCases(ctx context.Context, r *pb.ListTestCasesRequest) (*pb.ListTestCasesResponse, error) {
	if err := validateRequest(r); err != nil {
		return nil, err
	}
	deps := s.deps
	items, err := logic.ListResources(ctx, deps, r.GetOrgId(), r.GetProjectId(), "case", r.Search, r.FolderId, r.ApiId, "")
	if err != nil {
		return nil, err
	}
	return serializer.ListTestCases(map[string]any{"items": items})
}

func (s *Server) GetTestCase(ctx context.Context, r *pb.GetTestCaseRequest) (*pb.GetTestCaseResponse, error) {
	if err := validateRequest(r); err != nil {
		return nil, err
	}
	deps := s.deps
	items, err := logic.ListResources(ctx, deps, r.GetOrgId(), r.GetProjectId(), "case", "", nil, nil, r.Id)
	if err != nil {
		return nil, err
	}
	if len(items) != 1 {
		return nil, p.NotFound()
	}
	return serializer.GetTestCase(map[string]any{"resource": items[0]})
}

func (s *Server) CreateTestCase(ctx context.Context, r *pb.CreateTestCaseRequest) (*pb.CreateTestCaseResponse, error) {
	if err := validateRequest(r); err != nil {
		return nil, err
	}
	deps := s.deps
	in, err := resourceWriteInput(r, "case")
	if err != nil {
		return nil, err
	}
	item, err := logic.SaveResource(ctx, deps, in)
	if err != nil {
		return nil, err
	}
	return serializer.CreateTestCase(map[string]any{"resource": item})
}

func (s *Server) UpdateTestCase(ctx context.Context, r *pb.UpdateTestCaseRequest) (*pb.UpdateTestCaseResponse, error) {
	if err := validateRequest(r); err != nil {
		return nil, err
	}
	deps := s.deps
	in, err := resourceWriteInput(r, "case")
	if err != nil {
		return nil, err
	}
	item, err := logic.SaveResource(ctx, deps, in)
	if err != nil {
		return nil, err
	}
	return serializer.UpdateTestCase(map[string]any{"resource": item})
}

func (s *Server) DeleteTestCase(ctx context.Context, r *pb.DeleteTestCaseRequest) (*pb.DeleteTestCaseResponse, error) {
	if err := validateRequest(r); err != nil {
		return nil, err
	}
	deps := s.deps
	in := &logic.ResourceInput{OrgId: r.OrgId, ProjectId: r.ProjectId, Id: r.Id, Kind: "case"}
	if err := logic.DeleteResource(ctx, deps, in); err != nil {
		return nil, err
	}
	return serializer.DeleteTestCase(map[string]any{"ok": true})
}
