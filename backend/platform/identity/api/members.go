package api

import (
	"context"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/mi991203/TestMoon/backend/platform/identity/logic"
	"github.com/mi991203/TestMoon/backend/platform/identity/model"
	"github.com/mi991203/TestMoon/backend/platform/identity/serializer"
	p "github.com/mi991203/TestMoon/libs/go/platform"
	pb "github.com/mi991203/TestMoon/protocols/platform/identity/v1"
)

func (s *Server) ListOrganizationMembers(ctx context.Context, r *pb.ListOrganizationMembersRequest) (*pb.ListOrganizationMembersResponse, error) {
	if err := validateRequest(r); err != nil {
		return nil, err
	}
	deps := s.deps
	if e := p.RequireOrg(ctx, r.OrgId); e != nil {
		return nil, e
	}
	var members []model.Member
	if e := deps.DB.Where("org_id = ?", r.OrgId).Find(&members).Error; e != nil {
		return nil, p.DBError(e)
	}
	items := []*pb.OrganizationMember{}
	for _, m := range members {
		var u model.User
		if deps.DB.First(&u, "id = ?", m.UserID).Error == nil {
			items = append(items, serializer.Member(m, u))
		}
	}
	return &pb.ListOrganizationMembersResponse{Items: items}, nil
}

func (s *Server) SetOrganizationMember(ctx context.Context, r *pb.SetOrganizationMemberRequest) (*pb.SetOrganizationMemberResponse, error) {
	if err := validateRequest(r); err != nil {
		return nil, err
	}
	deps := s.deps
	if e := p.RequireOrgAdmin(ctx, r.OrgId); e != nil {
		return nil, e
	}
	var in struct {
		Username string `json:"username" validate:"required"`
		Role     string `json:"role" validate:"oneof=admin member"`
	}
	if e := p.Decode(r, &in); e != nil {
		return nil, e
	}
	var u model.User
	if deps.DB.First(&u, "username = ? AND disabled = ?", strings.ToLower(in.Username), false).Error != nil {
		return nil, p.Bad("有效账号不存在")
	}
	e := deps.DB.Transaction(func(tx *gorm.DB) error {
		var org model.Organization
		if e := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&org, "id = ?", r.OrgId).Error; e != nil {
			return e
		}
		var m model.Member
		existing := tx.Unscoped().First(&m, "org_id = ? AND user_id = ?", r.OrgId, u.ID).Error == nil
		if existing && m.DeletedAt == 0 && m.Role == "admin" && in.Role != "admin" {
			var n int64
			tx.Model(&model.Member{}).Where("org_id = ? AND role = ?", r.OrgId, "admin").Count(&n)
			if n <= 1 {
				return p.Bad("组织至少保留一个管理员")
			}
		}
		if in.Role == "remove" {
			if existing {
				return tx.Delete(&m).Error
			}
			return nil
		}
		if existing {
			return tx.Unscoped().Model(&m).Updates(map[string]any{"role": in.Role, "deleted_at": 0}).Error
		}
		return tx.Create(&model.Member{Base: p.NewBase(), OrgID: r.OrgId, UserID: u.ID, Role: in.Role}).Error
	})
	if e != nil {
		return nil, logic.Translate(e)
	}
	return &pb.SetOrganizationMemberResponse{Ok: true}, nil
}

func (s *Server) RemoveOrganizationMember(ctx context.Context, r *pb.RemoveOrganizationMemberRequest) (*pb.RemoveOrganizationMemberResponse, error) {
	if err := validateRequest(r); err != nil {
		return nil, err
	}
	deps := s.deps
	if e := p.RequireOrgAdmin(ctx, r.OrgId); e != nil {
		return nil, e
	}
	var in struct {
		Username string `json:"username" validate:"required"`
		Role     string `json:"role" validate:"oneof=admin member remove"`
	}
	in.Role = "remove"
	if e := p.Decode(r, &in); e != nil {
		return nil, e
	}
	var u model.User
	if deps.DB.First(&u, "username = ? AND disabled = ?", strings.ToLower(in.Username), false).Error != nil {
		return nil, p.Bad("有效账号不存在")
	}
	e := deps.DB.Transaction(func(tx *gorm.DB) error {
		var org model.Organization
		if e := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&org, "id = ?", r.OrgId).Error; e != nil {
			return e
		}
		var m model.Member
		existing := tx.Unscoped().First(&m, "org_id = ? AND user_id = ?", r.OrgId, u.ID).Error == nil
		if existing && m.DeletedAt == 0 && m.Role == "admin" && in.Role != "admin" {
			var n int64
			tx.Model(&model.Member{}).Where("org_id = ? AND role = ?", r.OrgId, "admin").Count(&n)
			if n <= 1 {
				return p.Bad("组织至少保留一个管理员")
			}
		}
		if in.Role == "remove" {
			if existing {
				return tx.Delete(&m).Error
			}
			return nil
		}
		if existing {
			return tx.Unscoped().Model(&m).Updates(map[string]any{"role": in.Role, "deleted_at": 0}).Error
		}
		return tx.Create(&model.Member{Base: p.NewBase(), OrgID: r.OrgId, UserID: u.ID, Role: in.Role}).Error
	})
	if e != nil {
		return nil, logic.Translate(e)
	}
	return &pb.RemoveOrganizationMemberResponse{Ok: true}, nil
}
