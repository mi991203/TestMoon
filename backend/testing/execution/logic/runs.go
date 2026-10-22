package logic

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"

	"github.com/mi991203/TestMoon/backend/testing/execution/dao"
	"github.com/mi991203/TestMoon/backend/testing/execution/internal/runtime"
	"github.com/mi991203/TestMoon/backend/testing/execution/model"
	p "github.com/mi991203/TestMoon/libs/go/platform"
	m "github.com/mi991203/TestMoon/libs/go/testing/domain"
	"github.com/mi991203/TestMoon/libs/go/transport"
	core "github.com/mi991203/TestMoon/protocols/testing/core/v1"
)

func StartRun(ctx context.Context, deps *runtime.Dependencies, r *RunInput) (model.Run, error) {
	in := struct {
		m.Task
		Debug          *m.Case
		IdempotencyKey string
	}{
		Task:  m.Task{Items: r.Items, EnvironmentID: r.EnvironmentID, RunnerID: r.RunnerID, TimeoutSeconds: r.TimeoutSeconds},
		Debug: r.Debug, IdempotencyKey: r.IdempotencyKey,
	}
	name := "手动执行"
	if r.Id != "" {
		var task model.Task
		if e := dao.WithProjectScope(deps.DB, r.OrgId, r.ProjectId).First(&task, "id = ?", r.Id).Error; e != nil {
			return model.Run{}, p.DBError(e)
		}
		definition, e := deps.Tasks.Definition(task)
		if e != nil {
			return model.Run{}, p.Bad("任务配置无效")
		}
		in.Task = definition
		name = task.Name
	}
	if in.RunnerID == "" {
		in.RunnerID = "public"
	}
	fingerprintBytes, _ := json.Marshal(map[string]any{"task": in.Task, "debug": in.Debug, "task_id": r.Id})
	fingerprint := p.Hash(string(fingerprintBytes))
	requestHash := p.Hash(r.OrgId + ":" + r.ProjectId + ":" + p.Actor(ctx).UserID + ":" + in.IdempotencyKey)
	var old model.Run
	if e := deps.DB.First(&old, "request_hash = ?", requestHash).Error; e == nil {
		if old.Fingerprint != fingerprint {
			return model.Run{}, p.Conflict("幂等键已用于不同请求")
		}
		return old, nil
	}
	query := deps.DB.Model(&model.Runner{}).Where("revoked = ? AND registered = ? AND last_seen > ?", false, true, time.Now().Add(-30*time.Second))
	if in.RunnerID == "public" {
		query = query.Where("public = ?", true)
	} else {
		query = query.Where("id = ? AND org_id = ? AND project_id = ? AND public = ?", in.RunnerID, r.OrgId, r.ProjectId, false)
	}
	var count int64
	if e := query.Count(&count).Error; e != nil {
		return model.Run{}, p.DBError(e)
	}
	if count == 0 {
		return model.Run{}, p.Conflict("所选Runner离线或不可用")
	}
	snapshotReq := jsonObject(in.Task)
	if in.Debug != nil {
		snapshotReq["debug"] = in.Debug
	}
	snapshotReq["org_id"] = r.OrgId
	snapshotReq["project_id"] = r.ProjectId
	snapshotRequest := &core.CreateSnapshotRequest{}
	if e := transport.Encode(snapshotReq, snapshotRequest); e != nil {
		return model.Run{}, e
	}
	out, e := deps.Core.Client.CreateSnapshot(p.Outgoing(ctx, deps.Config), snapshotRequest)
	if e != nil {
		return model.Run{}, e
	}
	var v struct {
		Snapshot m.Snapshot `json:"snapshot"`
	}
	if e = p.Unpack(out, &v); e != nil {
		return model.Run{}, e
	}
	raw, _ := json.Marshal(v.Snapshot)
	run := model.Run{Base: p.NewBase(), OrgID: r.OrgId, ProjectID: r.ProjectId, UserID: p.Actor(ctx).UserID, TaskID: r.Id, Name: name, State: "created", TargetRunnerID: in.RunnerID, RequestHash: requestHash, Fingerprint: fingerprint, Snapshot: p.Seal(deps.Config.EncryptionKey, string(raw)), TimeoutSeconds: v.Snapshot.TimeoutSeconds, QueueDeadline: time.Now().Add(time.Duration(deps.Config.QueueSeconds) * time.Second)}
	event := model.Outbox{Base: p.NewBase(), RunID: run.ID, LeaseUntil: time.Unix(1, 0).UTC()}
	payload, _ := json.Marshal(map[string]any{"event_id": event.ID, "version": 1, "run_id": run.ID})
	event.Payload = string(payload)
	e = deps.DB.Transaction(func(tx *gorm.DB) error {
		if e := tx.Create(&run).Error; e != nil {
			return e
		}
		return tx.Create(&event).Error
	})
	if e != nil {
		if deps.DB.First(&old, "request_hash = ?", requestHash).Error == nil && old.Fingerprint == fingerprint {
			return old, nil
		}
		return model.Run{}, p.DBError(e)
	}
	return run, nil
}

func ExecutionError(e error) error {
	if _, ok := status.FromError(e); ok {
		return e
	}
	return p.DBError(e)
}

func jsonObject(v any) map[string]any {
	b, _ := json.Marshal(v)
	var d map[string]any
	_ = json.Unmarshal(b, &d)
	return d
}

func AuthenticateRunner(ctx context.Context, deps *runtime.Dependencies) (model.Runner, error) {
	var node model.Runner
	auth := p.Header(ctx, "authorization")
	if !strings.HasPrefix(auth, "Bearer ") {
		return node, status.Error(codes.Unauthenticated, "缺少Runner凭证")
	}
	if deps.DB.First(&node, "credential_hash = ? AND registered = ? AND revoked = ?", p.Hash(strings.TrimPrefix(auth, "Bearer ")), true, false).Error != nil {
		return node, status.Error(codes.Unauthenticated, "Runner凭证无效")
	}
	return node, nil
}
