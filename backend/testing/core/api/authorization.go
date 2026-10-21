package api

import (
	"context"
	"strings"

	"github.com/mi991203/TestMoon/backend/testing/core/logic"
	"github.com/mi991203/TestMoon/backend/testing/core/serializer"
	p "github.com/mi991203/TestMoon/libs/go/platform"
	pb "github.com/mi991203/TestMoon/protocols/testing/core/v1"
)

func (s *Server) Authorize(ctx context.Context, r *pb.AuthorizeRequest) (*pb.AuthorizeResponse, error) {
	if err := validateRequest(r); err != nil {
		return nil, err
	}
	deps := s.deps
	permission := r.Permission
	if permission == "" {
		permission = "read"
	}
	if !strings.Contains("|read|write|manage|execute|active|", "|"+permission+"|") {
		return nil, p.Bad("无效权限")
	}
	_, role, e := logic.RequireProject(ctx, deps.DB, r.OrgId, r.ProjectId, permission)
	if e != nil {
		return nil, e
	}
	return serializer.Authorize(map[string]any{"role": role, "active": true})
}
