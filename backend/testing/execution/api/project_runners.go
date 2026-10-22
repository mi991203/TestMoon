package api

import (
	"context"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/mi991203/TestMoon/backend/testing/execution/dao"
	"github.com/mi991203/TestMoon/backend/testing/execution/logic"
	"github.com/mi991203/TestMoon/backend/testing/execution/model"
	"github.com/mi991203/TestMoon/backend/testing/execution/serializer"
	p "github.com/mi991203/TestMoon/libs/go/platform"
	pb "github.com/mi991203/TestMoon/protocols/testing/execution/v1"
)

func (s *Server) ListProjectRunners(ctx context.Context, r *pb.ListProjectRunnersRequest) (*pb.ListProjectRunnersResponse, error) {
	if err := validateRequest(r); err != nil {
		return nil, err
	}
	deps := s.deps
	if e := deps.Core.Authorize(ctx, r.OrgId, r.ProjectId, "read"); e != nil {
		return nil, e
	}
	var nodes []model.Runner
	if e := deps.DB.Where("((org_id = ? AND project_id = ?) OR public = ?) AND revoked = ?", r.OrgId, r.ProjectId, true, false).Order("created_at").Find(&nodes).Error; e != nil {
		return nil, p.DBError(e)
	}
	items := []any{}
	publicOnline := false
	publicCount := 0
	for _, node := range nodes {
		online := !node.Revoked && node.LastSeen != nil && time.Since(*node.LastSeen) < 30*time.Second
		if node.Public {
			publicCount++
			publicOnline = publicOnline || online
			continue
		}
		items = append(items, map[string]any{"id": node.ID, "name": node.Name, "online": online, "last_seen": node.LastSeen, "revoked": node.Revoked, "registered": node.Registered, "public": false, "concurrency": node.Concurrency, "version": node.Version})
	}
	items = append(items, map[string]any{"id": "public", "name": "公共Runner池", "online": publicOnline, "public": true, "nodes": publicCount})
	return serializer.ListProjectRunners(map[string]any{"items": items})
}

func (s *Server) CreatePrivateRunner(ctx context.Context, r *pb.CreatePrivateRunnerRequest) (*pb.CreatePrivateRunnerResponse, error) {
	if err := validateRequest(r); err != nil {
		return nil, err
	}
	concurrency, err := validateRunnerRegistration(r.Name, int(r.Concurrency))
	if err != nil {
		return nil, err
	}
	deps := s.deps
	if e := deps.Core.Authorize(ctx, r.OrgId, r.ProjectId, "manage"); e != nil {
		return nil, e
	}
	runner, token, e := logic.CreateRunner(ctx, deps, r.OrgId, r.ProjectId, r.Name, concurrency, false)
	if e != nil {
		return nil, e
	}
	return &pb.CreatePrivateRunnerResponse{Runner: serializer.Runner(runner), EnrollmentToken: token}, nil
}

func (s *Server) RevokePrivateRunner(ctx context.Context, r *pb.RevokePrivateRunnerRequest) (*pb.RevokePrivateRunnerResponse, error) {
	if err := validateRequest(r); err != nil {
		return nil, err
	}
	deps := s.deps
	if e := deps.Core.Authorize(ctx, r.OrgId, r.ProjectId, "manage"); e != nil {
		return nil, e
	}
	now := time.Now()
	e := deps.DB.Transaction(func(tx *gorm.DB) error {
		var node model.Runner
		if e := dao.WithProjectScope(tx, r.OrgId, r.ProjectId).Clauses(clause.Locking{Strength: "UPDATE"}).First(&node, "id = ? AND public = ?", r.Id, false).Error; e != nil {
			return e
		}
		if e := tx.Model(&node).Updates(map[string]any{"revoked": true, "credential_hash": "", "enrollment_hash": p.Hash(p.Token())}).Error; e != nil {
			return e
		}
		return tx.Model(&model.Run{}).Where("target_runner_id = ? AND state IN ?", node.ID, []string{"created", "queued", "running"}).Updates(map[string]any{"state": "error", "error": "Runner已撤销", "finished_at": now}).Error
	})
	if e != nil {
		return nil, logic.ExecutionError(e)
	}
	return serializer.RevokePrivateRunner(map[string]any{"ok": true})
}
