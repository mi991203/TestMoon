package api

import (
	"context"

	"github.com/mi991203/TestMoon/backend/testing/core/dao"
	"github.com/mi991203/TestMoon/backend/testing/core/logic"
	"github.com/mi991203/TestMoon/backend/testing/core/model"
	"github.com/mi991203/TestMoon/backend/testing/core/serializer"
	p "github.com/mi991203/TestMoon/libs/go/platform"
	pb "github.com/mi991203/TestMoon/protocols/testing/core/v1"
)

func (s *Server) ListProjectMembers(ctx context.Context, r *pb.ListProjectMembersRequest) (*pb.ListProjectMembersResponse, error) {
	if err := validateRequest(r); err != nil {
		return nil, err
	}
	deps := s.deps
	if _, _, e := logic.RequireProject(ctx, deps.DB, r.OrgId, r.ProjectId, "read"); e != nil {
		return nil, e
	}
	var items []model.Member
	if e := dao.WithProjectScope(deps.DB, r.OrgId, r.ProjectId).Find(&items).Error; e != nil {
		return nil, p.DBError(e)
	}
	return serializer.ListProjectMembers(map[string]any{"items": items})
}

func (s *Server) SetProjectMember(ctx context.Context, r *pb.SetProjectMemberRequest) (*pb.SetProjectMemberResponse, error) {
	if err := validateRequest(r); err != nil {
		return nil, err
	}
	deps := s.deps
	if r.Role != "admin" && r.Role != "editor" && r.Role != "viewer" {
		return nil, p.Bad("无效项目角色")
	}
	in := logic.MemberChange{OrgID: r.OrgId, ProjectID: r.ProjectId, UserID: r.UserId, Role: r.Role}
	if err := logic.SaveMember(ctx, deps, in); err != nil {
		return nil, err
	}
	return serializer.SetProjectMember(map[string]any{"ok": true})
}

func (s *Server) RemoveProjectMember(ctx context.Context, r *pb.RemoveProjectMemberRequest) (*pb.RemoveProjectMemberResponse, error) {
	if err := validateRequest(r); err != nil {
		return nil, err
	}
	deps := s.deps
	in := logic.MemberChange{OrgID: r.OrgId, ProjectID: r.ProjectId, UserID: r.UserId, Remove: true}
	if err := logic.SaveMember(ctx, deps, in); err != nil {
		return nil, err
	}
	return serializer.RemoveProjectMember(map[string]any{"ok": true})
}
