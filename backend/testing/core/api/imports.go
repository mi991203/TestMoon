package api

import (
	"context"
	"encoding/json"
	"time"

	"github.com/mi991203/TestMoon/backend/testing/core/internal/importer"
	"github.com/mi991203/TestMoon/backend/testing/core/logic"
	"github.com/mi991203/TestMoon/backend/testing/core/model"
	"github.com/mi991203/TestMoon/backend/testing/core/serializer"
	p "github.com/mi991203/TestMoon/libs/go/platform"
	pb "github.com/mi991203/TestMoon/protocols/testing/core/v1"
)

func (s *Server) PreviewImport(ctx context.Context, r *pb.PreviewImportRequest) (*pb.PreviewImportResponse, error) {
	if err := validateRequest(r); err != nil {
		return nil, err
	}
	deps := s.deps

	if _, _, e := logic.RequireProject(ctx, deps.DB, r.OrgId, r.ProjectId, "write"); e != nil {
		return nil, e
	}
	var in struct {
		Format  string `json:"format" validate:"oneof=postman openapi"`
		Content string `json:"content" validate:"required"`
	}
	if e := p.Decode(r, &in); e != nil {
		return nil, e
	}
	if len(in.Content) > 10<<20 {
		return nil, p.Bad("导入文件不能超过10MiB")
	}
	var plan model.ImportPlan
	var e error
	if in.Format == "postman" {
		plan, e = importer.Postman([]byte(in.Content))
	} else {
		plan, e = importer.OpenAPI([]byte(in.Content))
	}
	if e != nil {
		return nil, e
	}
	if len(plan.Entries) > 2000 {
		return nil, p.Bad("单次最多导入2000项")
	}
	if len(plan.Entries) == 0 {
		return nil, p.Bad("没有可导入的接口")
	}
	b, _ := json.Marshal(plan)
	batch := model.ImportBatch{Base: p.NewBase(), OrgID: r.OrgId, ProjectID: r.ProjectId, UserID: p.Actor(ctx).UserID, Plan: p.Seal(deps.Config.EncryptionKey, string(b)), ExpiresAt: time.Now().Add(time.Hour)}
	if e := deps.DB.Create(&batch).Error; e != nil {
		return nil, p.DBError(e)
	}
	return serializer.PreviewImport(map[string]any{"batch_id": batch.ID, "entries": plan.Entries, "diagnostics": plan.Diagnostics, "expires_at": batch.ExpiresAt})
}

func (s *Server) CommitImport(ctx context.Context, r *pb.CommitImportRequest) (*pb.CommitImportResponse, error) {
	if err := validateRequest(r); err != nil {
		return nil, err
	}
	deps := s.deps

	if _, _, e := logic.RequireProject(ctx, deps.DB, r.OrgId, r.ProjectId, "write"); e != nil {
		return nil, e
	}
	var in logic.ImportCommitInput
	if e := p.Decode(r, &in); e != nil {
		return nil, e
	}
	count, err := logic.CommitImport(ctx, deps, in)
	if err != nil {
		return nil, err
	}
	return serializer.CommitImport(map[string]any{"imported": count})
}
