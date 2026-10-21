package api

import (
	"context"
	"encoding/json"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/mi991203/TestMoon/backend/testing/execution/logic"
	"github.com/mi991203/TestMoon/backend/testing/execution/model"
	"github.com/mi991203/TestMoon/backend/testing/execution/serializer"
	p "github.com/mi991203/TestMoon/libs/go/platform"
	m "github.com/mi991203/TestMoon/libs/go/testing/domain"
	pb "github.com/mi991203/TestMoon/protocols/testing/execution/v1"
)

func (s *Server) RegisterRunner(ctx context.Context, r *pb.RegisterRunnerRequest) (*pb.RegisterRunnerResponse, error) {
	if err := validateRequest(r); err != nil {
		return nil, err
	}
	deps := s.deps

	var in struct {
		Token        string   `json:"token" validate:"required,max=128"`
		Version      string   `json:"version"`
		Capabilities []string `json:"capabilities"`
	}
	if e := p.Decode(r, &in); e != nil {
		return nil, e
	}
	supported := map[string]bool{}
	for _, x := range in.Capabilities {
		supported[x] = true
	}
	if !supported["http/1"] || !supported["javascript/1"] {
		return nil, p.Bad("Runner需要http/1和javascript/1能力")
	}
	credential := p.Token()
	var node model.Runner
	e := deps.DB.Transaction(func(tx *gorm.DB) error {
		if e := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&node, "enrollment_hash = ? AND revoked = ? AND registered = ?", p.Hash(in.Token), false, false).Error; e != nil {
			return e
		}
		caps, _ := json.Marshal(in.Capabilities)
		now := time.Now()
		return tx.Model(&node).Updates(map[string]any{"credential_hash": p.Hash(credential), "enrollment_hash": p.Hash(p.Token()), "registered": true, "last_seen": now, "version": in.Version, "capabilities": string(caps)}).Error
	})
	if e != nil {
		return nil, logic.ExecutionError(e)
	}
	return serializer.RegisterRunner(map[string]any{"runner_id": node.ID, "credential": credential, "public": node.Public, "concurrency": node.Concurrency})
}

func (s *Server) RunnerHeartbeat(ctx context.Context, r *pb.RunnerHeartbeatRequest) (*pb.RunnerHeartbeatResponse, error) {
	if err := validateRequest(r); err != nil {
		return nil, err
	}
	deps := s.deps
	node, e := logic.AuthenticateRunner(ctx, deps)
	if e != nil {
		return nil, e
	}
	var in struct {
		Leases map[string]string `json:"leases"`
	}
	if e := p.Decode(r, &in); e != nil {
		return nil, e
	}
	now := time.Now()
	if e := deps.DB.Model(&node).Update("last_seen", now).Error; e != nil {
		return nil, p.DBError(e)
	}
	active := map[string]bool{}
	for id, lease := range in.Leases {
		result := deps.DB.Model(&model.Run{}).Where("id = ? AND node_id = ? AND lease_hash = ? AND state = ? AND lease_until > ? AND deadline > ?", id, node.ID, p.Hash(lease), "running", now, now).Update("lease_until", now.Add(45*time.Second))
		if result.Error != nil {
			return nil, p.DBError(result.Error)
		}
		active[id] = result.RowsAffected == 1
	}
	return serializer.RunnerHeartbeat(map[string]any{"active": active})
}

func (s *Server) PollRunnerJob(ctx context.Context, r *pb.PollRunnerJobRequest) (*pb.PollRunnerJobResponse, error) {
	if err := validateRequest(r); err != nil {
		return nil, err
	}
	deps := s.deps
	node, e := logic.AuthenticateRunner(ctx, deps)
	if e != nil {
		return nil, e
	}
	job, err := logic.PollJob(ctx, deps, node)
	if err != nil {
		return nil, err
	}
	return serializer.PollRunnerJob(map[string]any{"job": job})
}

func (s *Server) ProgressRun(ctx context.Context, r *pb.ProgressRunRequest) (*pb.ProgressRunResponse, error) {
	if err := validateRequest(r); err != nil {
		return nil, err
	}
	deps := s.deps
	node, e := logic.AuthenticateRunner(ctx, deps)
	if e != nil {
		return nil, e
	}
	var in struct {
		Lease  string   `json:"lease" validate:"required"`
		Result m.Result `json:"result"`
	}
	if e := p.Decode(r, &in); e != nil {
		return nil, e
	}
	var run model.Run
	now := time.Now()
	if e := deps.DB.First(&run, "id = ? AND node_id = ? AND lease_hash = ? AND state = ? AND lease_until > ? AND deadline > ?", r.Id, node.ID, p.Hash(in.Lease), "running", now, now).Error; e != nil {
		return nil, p.Conflict("执行凭证已过期或执行已结束")
	}
	if !node.Public && (run.OrgID != node.OrgID || run.ProjectID != node.ProjectID) {
		return nil, p.Denied()
	}
	if e := deps.Core.Authorize(ctx, run.OrgID, run.ProjectID, "active"); e != nil {
		return nil, e
	}
	safe, e := logic.SanitizeResult(deps, run, in.Result)
	if e != nil {
		return nil, e
	}
	b, _ := json.Marshal(safe)
	if len(b) > 8<<20 {
		return nil, p.Bad("报告超过8MiB")
	}
	updates := map[string]any{"result": p.Seal(deps.Config.EncryptionKey, string(b))}

	result := deps.DB.Model(&model.Run{}).Where("id = ? AND node_id = ? AND lease_hash = ? AND state = ?", run.ID, node.ID, p.Hash(in.Lease), "running").Updates(updates)
	if result.Error != nil {
		return nil, p.DBError(result.Error)
	}
	if result.RowsAffected != 1 {
		return nil, p.Conflict("执行已结束")
	}
	return serializer.ProgressRun(map[string]any{"ok": true})
}

func (s *Server) CompleteRun(ctx context.Context, r *pb.CompleteRunRequest) (*pb.CompleteRunResponse, error) {
	if err := validateRequest(r); err != nil {
		return nil, err
	}
	deps := s.deps
	node, e := logic.AuthenticateRunner(ctx, deps)
	if e != nil {
		return nil, e
	}
	var in struct {
		Lease  string   `json:"lease" validate:"required"`
		Result m.Result `json:"result"`
	}
	if e := p.Decode(r, &in); e != nil {
		return nil, e
	}
	var run model.Run
	now := time.Now()
	if e := deps.DB.First(&run, "id = ? AND node_id = ? AND lease_hash = ? AND state = ? AND lease_until > ? AND deadline > ?", r.Id, node.ID, p.Hash(in.Lease), "running", now, now).Error; e != nil {
		return nil, p.Conflict("执行凭证已过期或执行已结束")
	}
	if !node.Public && (run.OrgID != node.OrgID || run.ProjectID != node.ProjectID) {
		return nil, p.Denied()
	}
	if e := deps.Core.Authorize(ctx, run.OrgID, run.ProjectID, "active"); e != nil {
		return nil, e
	}
	safe, e := logic.SanitizeResult(deps, run, in.Result)
	if e != nil {
		return nil, e
	}
	b, _ := json.Marshal(safe)
	if len(b) > 8<<20 {
		return nil, p.Bad("报告超过8MiB")
	}
	updates := map[string]any{"result": p.Seal(deps.Config.EncryptionKey, string(b))}

	if !logic.ValidResult(in.Result) {
		return nil, p.Bad("无效执行结果或汇总不一致")
	}
	updates["state"] = in.Result.Status
	updates["finished_at"] = now

	result := deps.DB.Model(&model.Run{}).Where("id = ? AND node_id = ? AND lease_hash = ? AND state = ?", run.ID, node.ID, p.Hash(in.Lease), "running").Updates(updates)
	if result.Error != nil {
		return nil, p.DBError(result.Error)
	}
	if result.RowsAffected != 1 {
		return nil, p.Conflict("执行已结束")
	}
	return serializer.CompleteRun(map[string]any{"ok": true})
}
