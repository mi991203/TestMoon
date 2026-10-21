package logic

import (
	"context"
	"encoding/json"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"

	"github.com/mi991203/TestMoon/backend/testing/core/internal/runtime"
	"github.com/mi991203/TestMoon/backend/testing/core/model"
	p "github.com/mi991203/TestMoon/libs/go/platform"
)

func RequireProject(ctx context.Context, db *gorm.DB, org, id, permission string) (model.Project, string, error) {
	var project model.Project
	if org == "" || id == "" {
		return project, "", p.Bad("缺少组织或项目")
	}
	if e := db.First(&project, "org_id = ? AND id = ?", org, id).Error; e != nil {
		return project, "", p.DBError(e)
	}
	if permission == "active" && p.Actor(ctx).UserID == "" {
		return project, "", nil
	}
	if e := p.RequireOrg(ctx, org); e != nil {
		return project, "", e
	}
	role := ""
	if p.Actor(ctx).Organizations[org] == "admin" {
		role = "admin"
	} else {
		var member model.Member
		if db.First(&member, "org_id = ? AND project_id = ? AND user_id = ?", org, id, p.Actor(ctx).UserID).Error == nil {
			role = member.Role
		}
	}
	if permission == "active" {
		permission = "read"
	}
	allowed := role == "admin" || ((role == "editor" || role == "viewer") && permission == "read") || (role == "editor" && (permission == "write" || permission == "execute"))
	if !allowed {
		return project, role, p.Denied()
	}
	return project, role, nil
}

func readResourceDocument(deps *runtime.Dependencies, a model.Resource) (map[string]any, error) {
	if a.Document == "" {
		return deps.Assets.Document(deps.DB, a)
	}
	raw, e := p.Open(deps.Config.EncryptionKey, a.Document)
	if e != nil {
		return nil, status.Error(codes.Internal, "资源解密失败")
	}
	var doc map[string]any
	if json.Unmarshal([]byte(raw), &doc) != nil {
		return nil, status.Error(codes.Internal, "资源格式无效")
	}
	return doc, nil
}

func LoadResource(deps *runtime.Dependencies, db *gorm.DB, org, project, id, kind string) (model.Resource, error) {
	var a model.Resource
	e := db.Transaction(func(tx *gorm.DB) error {
		q := tx.Where("org_id = ? AND project_id = ? AND id = ?", org, project, id)
		if kind != "" {
			q = q.Where("kind = ?", kind)
		}
		if err := q.First(&a).Error; err != nil {
			return err
		}
		return deps.Assets.Hydrate(tx, &a)
	})
	return a, p.DBError(e)
}

func resourceView(deps *runtime.Dependencies, a model.Resource, role string) (map[string]any, error) {
	doc, e := readResourceDocument(deps, a)
	if e != nil {
		return nil, e
	}
	if role == "viewer" {
		doc = p.Sanitize(doc, nil).(map[string]any)
	}
	return map[string]any{"id": a.ID, "name": a.Name, "kind": a.Kind, "folder_id": a.FolderID, "revision": a.Revision, "updated_at": a.UpdatedAt, "created_at": a.CreatedAt, "org_id": a.OrgID, "project_id": a.ProjectID, "document": doc}, nil
}
