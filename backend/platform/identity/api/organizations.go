package api

import (
	"context"

	"gorm.io/gorm"

	"github.com/mi991203/TestMoon/backend/platform/identity/logic"
	"github.com/mi991203/TestMoon/backend/platform/identity/model"
	"github.com/mi991203/TestMoon/backend/platform/identity/serializer"
	p "github.com/mi991203/TestMoon/libs/go/platform"
	pb "github.com/mi991203/TestMoon/protocols/platform/identity/v1"
)

func (s *Server) ListOrganizations(ctx context.Context, r *pb.ListOrganizationsRequest) (*pb.ListOrganizationsResponse, error) {
	if err := validateRequest(r); err != nil {
		return nil, err
	}
	deps := s.deps
	actor := p.Actor(ctx)
	q := deps.DB.Model(&model.Organization{})
	if !actor.Admin {
		ids := []string{}
		for id := range actor.Organizations {
			ids = append(ids, id)
		}
		q = q.Where("id IN ?", ids)
	}
	var items []model.Organization
	if e := q.Order("created_at").Find(&items).Error; e != nil {
		return nil, p.DBError(e)
	}
	return &pb.ListOrganizationsResponse{Items: serializer.Organizations(items)}, nil
}

func (s *Server) CreateOrganization(ctx context.Context, r *pb.CreateOrganizationRequest) (*pb.CreateOrganizationResponse, error) {
	if err := validateRequest(r); err != nil {
		return nil, err
	}
	deps := s.deps
	if e := p.RequireAdmin(ctx); e != nil {
		return nil, e
	}
	var in struct {
		Name        string `json:"name" validate:"required,max=128"`
		AdminUserID string `json:"admin_user_id"`
	}
	if e := p.Decode(r, &in); e != nil {
		return nil, e
	}
	o := model.Organization{Base: p.NewBase(), Name: in.Name}
	e := deps.DB.Transaction(func(tx *gorm.DB) error {

		var u model.User
		if tx.First(&u, "id = ? AND disabled = ?", in.AdminUserID, false).Error != nil {
			return p.Bad("请选择有效的组织管理员账号")
		}
		if e := tx.Create(&o).Error; e != nil {
			return e
		}
		return tx.Create(&model.Member{Base: p.NewBase(), OrgID: o.ID, UserID: u.ID, Role: "admin"}).Error
	})
	if e != nil {
		return nil, logic.Translate(e)
	}
	return &pb.CreateOrganizationResponse{Organization: serializer.Organization(o)}, nil
}

func (s *Server) UpdateOrganization(ctx context.Context, r *pb.UpdateOrganizationRequest) (*pb.UpdateOrganizationResponse, error) {
	if err := validateRequest(r); err != nil {
		return nil, err
	}
	deps := s.deps
	if e := p.RequireAdmin(ctx); e != nil {
		return nil, e
	}
	var in struct {
		Name        string `json:"name" validate:"required,max=128"`
		AdminUserID string `json:"admin_user_id"`
	}
	if e := p.Decode(r, &in); e != nil {
		return nil, e
	}
	o := model.Organization{Base: p.NewBase(), Name: in.Name}
	e := deps.DB.Transaction(func(tx *gorm.DB) error {
		if e := tx.First(&o, "id = ?", r.Id).Error; e != nil {
			return e
		}
		return tx.Model(&o).Update("name", in.Name).Error
	})
	if e != nil {
		return nil, logic.Translate(e)
	}
	return &pb.UpdateOrganizationResponse{Organization: serializer.Organization(o)}, nil
}
