package api

import (
	"context"
	"encoding/json"
	"time"

	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/mi991203/TestMoon/backend/testing/core/dao"
	"github.com/mi991203/TestMoon/backend/testing/core/logic"
	"github.com/mi991203/TestMoon/backend/testing/core/model"
	"github.com/mi991203/TestMoon/backend/testing/core/serializer"
	p "github.com/mi991203/TestMoon/libs/go/platform"
	m "github.com/mi991203/TestMoon/libs/go/testing/domain"
	pb "github.com/mi991203/TestMoon/protocols/testing/core/v1"
)

func (s *Server) ListShares(ctx context.Context, r *pb.ListSharesRequest) (*pb.ListSharesResponse, error) {
	if err := validateRequest(r); err != nil {
		return nil, err
	}
	deps := s.deps
	permission := "read"
	_, role, e := logic.RequireProject(ctx, deps.DB, r.OrgId, r.ProjectId, permission)
	if e != nil {
		return nil, e
	}
	var all []model.Share
	if e := dao.WithProjectScope(deps.DB, r.OrgId, r.ProjectId).Order("created_at DESC").Find(&all).Error; e != nil {
		return nil, p.DBError(e)
	}
	items := []any{}
	for _, v := range all {
		item := map[string]any{"id": v.ID, "name": v.Name, "expires_at": v.ExpiresAt, "password_required": v.Password != ""}
		if role != "viewer" {
			token, e := p.Open(deps.Config.EncryptionKey, v.Token)
			if e != nil {
				return nil, status.Error(codes.Internal, "分享链接解密失败")
			}
			item["url"] = deps.Config.WebOrigin + "/share/" + token
		}
		items = append(items, item)
	}
	return serializer.ListShares(map[string]any{"items": items})
}

func (s *Server) CreateShare(ctx context.Context, r *pb.CreateShareRequest) (*pb.CreateShareResponse, error) {
	if err := validateRequest(r); err != nil {
		return nil, err
	}
	deps := s.deps
	permission := "write"
	_, _, e := logic.RequireProject(ctx, deps.DB, r.OrgId, r.ProjectId, permission)
	if e != nil {
		return nil, e
	}
	var in struct {
		Name      string     `json:"name" validate:"required,max=128"`
		APIIDs    []string   `json:"api_ids" validate:"required,min=1,max=100"`
		Password  string     `json:"password" validate:"max=72"`
		ExpiresAt *time.Time `json:"expires_at"`
	}
	if e := p.Decode(r, &in); e != nil {
		return nil, e
	}
	for _, id := range in.APIIDs {
		if _, e := logic.LoadResource(deps, deps.DB, r.OrgId, r.ProjectId, id, "api"); e != nil {
			return nil, e
		}
	}
	if in.ExpiresAt != nil && !in.ExpiresAt.After(time.Now()) {
		return nil, p.Bad("有效期应晚于当前时间")
	}
	token := p.Token()
	ids, _ := json.Marshal(in.APIIDs)
	v := model.Share{Base: p.NewBase(), OrgID: r.OrgId, ProjectID: r.ProjectId, Name: in.Name, TokenHash: p.Hash(token), Token: p.Seal(deps.Config.EncryptionKey, token), APIIDs: string(ids), ExpiresAt: in.ExpiresAt}
	if in.Password != "" {
		h, e := bcrypt.GenerateFromPassword([]byte(in.Password), 12)
		if e != nil {
			return nil, p.Bad("无效分享密码")
		}
		v.Password = string(h)
	}
	if e := deps.DB.Create(&v).Error; e != nil {
		return nil, p.DBError(e)
	}
	return serializer.CreateShare(map[string]any{"id": v.ID, "url": deps.Config.WebOrigin + "/share/" + token})
}

func (s *Server) DeleteShare(ctx context.Context, r *pb.DeleteShareRequest) (*pb.DeleteShareResponse, error) {
	if err := validateRequest(r); err != nil {
		return nil, err
	}
	deps := s.deps
	permission := "write"
	_, _, e := logic.RequireProject(ctx, deps.DB, r.OrgId, r.ProjectId, permission)
	if e != nil {
		return nil, e
	}
	result := dao.WithProjectScope(deps.DB, r.OrgId, r.ProjectId).Where("id = ?", r.Id).Delete(&model.Share{})
	if result.Error != nil {
		return nil, p.DBError(result.Error)
	}
	return serializer.DeleteShare(map[string]any{"ok": true})
}

func (s *Server) ReadSharedDocument(ctx context.Context, r *pb.ReadSharedDocumentRequest) (*pb.ReadSharedDocumentResponse, error) {
	if err := validateRequest(r); err != nil {
		return nil, err
	}
	deps := s.deps

	var in struct {
		Token    string `json:"token" validate:"required,max=128"`
		Password string `json:"password" validate:"max=72"`
	}
	if e := p.Decode(r, &in); e != nil {
		return nil, e
	}
	var share model.Share
	if e := deps.DB.First(&share, "token_hash = ?", p.Hash(in.Token)).Error; e != nil {
		return nil, p.NotFound()
	}
	if share.ExpiresAt != nil && !share.ExpiresAt.After(time.Now()) {
		return nil, p.NotFound()
	}
	var project model.Project
	if deps.DB.First(&project, "id = ? AND org_id = ?", share.ProjectID, share.OrgID).Error != nil {
		return nil, p.NotFound()
	}
	if share.Password != "" {
		key := "share_attempt:" + share.TokenHash
		n, e := deps.Redis.Get(ctx, key).Int()
		if e != nil && e.Error() != "redis: nil" {
			return nil, status.Error(codes.Unavailable, "分享校验暂不可用")
		}
		if n >= 20 {
			return nil, status.Error(codes.ResourceExhausted, "尝试次数过多")
		}
		if bcrypt.CompareHashAndPassword([]byte(share.Password), []byte(in.Password)) != nil {
			n, _ := deps.Redis.Incr(ctx, key).Result()
			if n == 1 {
				_ = deps.Redis.Expire(ctx, key, 5*time.Minute).Err()
			}
			return nil, status.Error(codes.Unauthenticated, "需要正确的分享密码")
		}
		_ = deps.Redis.Del(ctx, key).Err()
	}
	var ids []string
	_ = json.Unmarshal([]byte(share.APIIDs), &ids)
	items := []any{}
	for _, id := range ids {
		a, e := logic.LoadResource(deps, deps.DB, share.OrgID, share.ProjectID, id, "api")
		if e != nil {
			continue
		}
		api, e := logic.DecodeDocument[m.API](deps, a)
		if e != nil {
			return nil, e
		}
		api.Request.Auth = m.Auth{}
		items = append(items, map[string]any{"id": a.ID, "name": a.Name, "document": p.Sanitize(api, nil)})
	}
	return serializer.ReadSharedDocument(map[string]any{"name": share.Name, "items": items})
}
