package api

import (
	"context"

	"github.com/mi991203/TestMoon/backend/testing/core/logic"
	"github.com/mi991203/TestMoon/backend/testing/core/serializer"
	p "github.com/mi991203/TestMoon/libs/go/platform"
	pb "github.com/mi991203/TestMoon/protocols/testing/core/v1"
)

func (s *Server) CreateSnapshot(ctx context.Context, r *pb.CreateSnapshotRequest) (*pb.CreateSnapshotResponse, error) {
	if err := validateRequest(r); err != nil {
		return nil, err
	}
	deps := s.deps

	if _, _, e := logic.RequireProject(ctx, deps.DB, r.OrgId, r.ProjectId, "execute"); e != nil {
		return nil, e
	}
	var in logic.SnapshotInput
	if e := p.Decode(r, &in); e != nil {
		return nil, e
	}
	out, err := logic.BuildSnapshot(ctx, deps, in)
	if err != nil {
		return nil, err
	}
	return serializer.CreateSnapshot(map[string]any{"snapshot": out})
}
