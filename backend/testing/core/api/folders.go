package api

import (
	"context"

	"github.com/mi991203/TestMoon/backend/testing/core/logic"
	"github.com/mi991203/TestMoon/backend/testing/core/serializer"
	p "github.com/mi991203/TestMoon/libs/go/platform"
	pb "github.com/mi991203/TestMoon/protocols/testing/core/v1"
)

func (s *Server) ListFolders(ctx context.Context, r *pb.ListFoldersRequest) (*pb.ListFoldersResponse, error) {
	if err := validateRequest(r); err != nil {
		return nil, err
	}
	deps := s.deps
	items, err := logic.ListResources(ctx, deps, r.GetOrgId(), r.GetProjectId(), "folder", r.Search, r.FolderId, nil, "")
	if err != nil {
		return nil, err
	}
	return serializer.ListFolders(map[string]any{"items": items})
}

func (s *Server) GetFolder(ctx context.Context, r *pb.GetFolderRequest) (*pb.GetFolderResponse, error) {
	if err := validateRequest(r); err != nil {
		return nil, err
	}
	deps := s.deps
	items, err := logic.ListResources(ctx, deps, r.GetOrgId(), r.GetProjectId(), "folder", "", nil, nil, r.Id)
	if err != nil {
		return nil, err
	}
	if len(items) != 1 {
		return nil, p.NotFound()
	}
	return serializer.GetFolder(map[string]any{"resource": items[0]})
}

func (s *Server) CreateFolder(ctx context.Context, r *pb.CreateFolderRequest) (*pb.CreateFolderResponse, error) {
	if err := validateRequest(r); err != nil {
		return nil, err
	}
	deps := s.deps
	in, err := resourceWriteInput(r, "folder")
	if err != nil {
		return nil, err
	}
	item, err := logic.SaveResource(ctx, deps, in)
	if err != nil {
		return nil, err
	}
	return serializer.CreateFolder(map[string]any{"resource": item})
}

func (s *Server) UpdateFolder(ctx context.Context, r *pb.UpdateFolderRequest) (*pb.UpdateFolderResponse, error) {
	if err := validateRequest(r); err != nil {
		return nil, err
	}
	deps := s.deps
	in, err := resourceWriteInput(r, "folder")
	if err != nil {
		return nil, err
	}
	item, err := logic.SaveResource(ctx, deps, in)
	if err != nil {
		return nil, err
	}
	return serializer.UpdateFolder(map[string]any{"resource": item})
}

func (s *Server) DeleteFolder(ctx context.Context, r *pb.DeleteFolderRequest) (*pb.DeleteFolderResponse, error) {
	if err := validateRequest(r); err != nil {
		return nil, err
	}
	deps := s.deps
	in := &logic.ResourceInput{OrgId: r.OrgId, ProjectId: r.ProjectId, Id: r.Id, Kind: "folder"}
	if err := logic.DeleteResource(ctx, deps, in); err != nil {
		return nil, err
	}
	return serializer.DeleteFolder(map[string]any{"ok": true})
}
