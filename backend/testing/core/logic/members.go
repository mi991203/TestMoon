package logic

import (
	"context"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/mi991203/TestMoon/backend/testing/core/internal/runtime"
	"github.com/mi991203/TestMoon/backend/testing/core/model"
	p "github.com/mi991203/TestMoon/libs/go/platform"
)

type MemberChange struct {
	OrgID, ProjectID, UserID, Role string
	Remove                         bool
}

func SaveMember(ctx context.Context, deps *runtime.Dependencies, r MemberChange) error {
	if _, _, e := RequireProject(ctx, deps.DB, r.OrgID, r.ProjectID, "manage"); e != nil {
		return e
	}
	in := r
	if in.UserID == "" {
		return p.Bad("缺少成员ID")
	}
	if e := VerifyOrgMember(ctx, deps, r.OrgID, in.UserID); e != nil {
		return e
	}
	e := deps.DB.Transaction(func(tx *gorm.DB) error {
		var pr model.Project
		if e := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&pr, "org_id = ? AND id = ?", r.OrgID, r.ProjectID).Error; e != nil {
			return e
		}
		var old model.Member
		found := tx.Unscoped().Where("org_id = ? AND project_id = ?", r.OrgID, r.ProjectID).First(&old, "user_id = ?", in.UserID).Error == nil
		if in.Remove {
			if found {
				return tx.Delete(&old).Error
			}
			return nil
		}
		if found {
			return tx.Unscoped().Model(&old).Updates(map[string]any{"role": in.Role, "deleted_at": 0}).Error
		}
		return tx.Create(&model.Member{Base: p.NewBase(), OrgID: r.OrgID, ProjectID: r.ProjectID, UserID: in.UserID, Role: in.Role}).Error
	})
	if e != nil {
		return CoreError(e)
	}
	return nil
}
