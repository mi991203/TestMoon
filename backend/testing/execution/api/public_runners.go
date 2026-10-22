package api

import (
	"context"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/mi991203/TestMoon/backend/testing/execution/logic"
	"github.com/mi991203/TestMoon/backend/testing/execution/model"
	"github.com/mi991203/TestMoon/backend/testing/execution/serializer"
	p "github.com/mi991203/TestMoon/libs/go/platform"
	pb "github.com/mi991203/TestMoon/protocols/testing/execution/v1"
)

func (s *Server) ListPublicRunners(ctx context.Context, r *pb.ListPublicRunnersRequest) (*pb.ListPublicRunnersResponse, error) {
	if err := validateRequest(r); err != nil {
		return nil, err
	}
	deps := s.deps
	if e := p.RequireAdmin(ctx); e != nil {
		return nil, e
	}
	var nodes []model.Runner
	if e := deps.DB.Where("public = ?", true).Order("created_at").Find(&nodes).Error; e != nil {
		return nil, p.DBError(e)
	}
	items := []any{}
	for _, node := range nodes {
		items = append(items, map[string]any{"id": node.ID, "name": node.Name, "public": true, "revoked": node.Revoked, "concurrency": node.Concurrency, "online": !node.Revoked && node.LastSeen != nil && time.Since(*node.LastSeen) < 30*time.Second})
	}
	return serializer.ListPublicRunners(map[string]any{"items": items})
}

func (s *Server) CreatePublicRunner(ctx context.Context, r *pb.CreatePublicRunnerRequest) (*pb.CreatePublicRunnerResponse, error) {
	if err := validateRequest(r); err != nil {
		return nil, err
	}
	concurrency, err := validateRunnerRegistration(r.Name, int(r.Concurrency))
	if err != nil {
		return nil, err
	}
	deps := s.deps
	runner, token, e := logic.CreateRunner(ctx, deps, "", "", r.Name, concurrency, true)
	if e != nil {
		return nil, e
	}
	return &pb.CreatePublicRunnerResponse{Runner: serializer.Runner(runner), EnrollmentToken: token}, nil
}

func (s *Server) RevokePublicRunner(ctx context.Context, r *pb.RevokePublicRunnerRequest) (*pb.RevokePublicRunnerResponse, error) {
	if err := validateRequest(r); err != nil {
		return nil, err
	}
	deps := s.deps
	if e := p.RequireAdmin(ctx); e != nil {
		return nil, e
	}
	e := deps.DB.Transaction(func(tx *gorm.DB) error {
		var node model.Runner
		if e := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&node, "id = ? AND public = ?", r.Id, true).Error; e != nil {
			return e
		}
		if e := tx.Model(&node).Updates(map[string]any{"revoked": true, "credential_hash": "", "enrollment_hash": p.Hash(p.Token())}).Error; e != nil {
			return e
		}
		return tx.Model(&model.Run{}).Where("node_id = ? AND state = ?", node.ID, "running").Updates(map[string]any{"state": "error", "error": "Runner已撤销", "finished_at": time.Now()}).Error
	})
	if e != nil {
		return nil, logic.ExecutionError(e)
	}
	return serializer.RevokePublicRunner(map[string]any{"ok": true})
}
