package logic

import (
	"context"
	"encoding/json"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/mi991203/TestMoon/backend/testing/execution/internal/runtime"
	"github.com/mi991203/TestMoon/backend/testing/execution/model"
	p "github.com/mi991203/TestMoon/libs/go/platform"
	m "github.com/mi991203/TestMoon/libs/go/testing/domain"
)

// RunnerJob is the leased execution payload consumed by a Runner.
type RunnerJob struct {
	RunID    string     `json:"run_id"`
	Lease    string     `json:"lease"`
	Public   bool       `json:"public"`
	Snapshot m.Snapshot `json:"snapshot"`
}

func PollJob(ctx context.Context, deps *runtime.Dependencies, node model.Runner) (*RunnerJob, error) {
	var selected model.Run
	lease := p.Token()
	e := deps.DB.Transaction(func(tx *gorm.DB) error {
		var locked model.Runner
		if e := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&locked, "id = ? AND revoked = ?", node.ID, false).Error; e != nil {
			return e
		}
		if locked.LastSeen == nil || time.Since(*locked.LastSeen) > 30*time.Second {
			return p.Conflict("Runner心跳过期")
		}
		var busy int64
		if e := tx.Model(&model.Run{}).Where("node_id = ? AND state = ?", node.ID, "running").Count(&busy).Error; e != nil {
			return e
		}
		if busy >= int64(node.Concurrency) {
			return nil
		}
		q := tx.Where("state = ? AND queue_deadline > ?", "queued", time.Now())
		if node.Public {
			q = q.Where("target_runner_id = ?", "public")
		} else {
			q = q.Where("target_runner_id = ? AND org_id = ? AND project_id = ?", node.ID, node.OrgID, node.ProjectID)
		}
		var candidates []model.Run
		if e := q.Clauses(clause.Locking{Strength: "UPDATE", Options: "SKIP LOCKED"}).Order("created_at,id").Limit(20).Find(&candidates).Error; e != nil {
			return e
		}
		for _, run := range candidates {
			if e := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&model.Budget{OrgID: run.OrgID}).Error; e != nil {
				return e
			}
			var budget model.Budget
			if e := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&budget, "org_id = ?", run.OrgID).Error; e != nil {
				return e
			}
			var orgBusy int64
			if e := tx.Model(&model.Run{}).Where("org_id = ? AND state = ?", run.OrgID, "running").Count(&orgBusy).Error; e != nil {
				return e
			}
			if orgBusy >= int64(deps.Config.OrgConcurrency) {
				continue
			}
			now := time.Now()
			deadline := now.Add(time.Duration(run.TimeoutSeconds) * time.Second)
			until := now.Add(45 * time.Second)
			res := tx.Model(&model.Run{}).Where("id = ? AND state = ?", run.ID, "queued").Updates(map[string]any{"state": "running", "node_id": node.ID, "lease_hash": p.Hash(lease), "started_at": now, "deadline": deadline, "lease_until": until})
			if res.Error != nil {
				return res.Error
			}
			if res.RowsAffected == 1 {
				selected = run
				selected.State = "running"
				break
			}
		}
		return nil
	})
	if e != nil {
		return nil, ExecutionError(e)
	}
	if selected.ID == "" {
		return nil, nil
	}
	if e := deps.Core.Authorize(ctx, selected.OrgID, selected.ProjectID, "active"); e != nil {
		deps.DB.Model(&model.Run{}).Where("id = ? AND state = ?", selected.ID, "running").Updates(map[string]any{"state": "error", "error": "项目不可用", "finished_at": time.Now()})
		return nil, e
	}
	raw, e := p.Open(deps.Config.EncryptionKey, selected.Snapshot)
	if e != nil {
		return nil, e
	}
	var snapshot m.Snapshot
	if e = json.Unmarshal([]byte(raw), &snapshot); e != nil {
		return nil, e
	}
	return &RunnerJob{RunID: selected.ID, Lease: lease, Public: node.Public, Snapshot: snapshot}, nil
}

func ValidResult(v m.Result) bool {
	switch v.Status {
	case "passed", "failed", "unverified", "error", "timed_out", "canceled":
	default:
		return false
	}
	if len(v.Items) == 0 {
		return v.Status == "error" || v.Status == "timed_out" || v.Status == "canceled"
	}
	if v.Status == "passed" {
		for _, item := range v.Items {
			if item.Status != "passed" {
				return false
			}
			for _, step := range item.Steps {
				if step.Status != "passed" || len(step.Assertions) == 0 {
					return false
				}
				for _, a := range step.Assertions {
					if !a.Passed {
						return false
					}
				}
			}
		}
	}
	return true
}

func SanitizeResult(deps *runtime.Dependencies, run model.Run, result m.Result) (any, error) {
	raw, e := p.Open(deps.Config.EncryptionKey, run.Snapshot)
	if e != nil {
		return nil, e
	}
	var snap m.Snapshot
	if e = json.Unmarshal([]byte(raw), &snap); e != nil {
		return nil, e
	}
	secrets := []string{}
	for _, v := range snap.Environment.Variables {
		if v.Secret && v.Value != "" {
			secrets = append(secrets, v.Value)
		}
	}
	for _, item := range snap.Items {
		for _, st := range item.Steps {
			a := st.Case.Request.Auth
			for _, x := range []string{a.Password, a.Token, a.Value} {
				if x != "" {
					secrets = append(secrets, x)
				}
			}
		}
	}
	return p.Sanitize(result, secrets), nil
}
