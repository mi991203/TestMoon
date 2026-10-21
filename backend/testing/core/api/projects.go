package api

import (
	"context"

	"gorm.io/gorm"

	"github.com/mi991203/TestMoon/backend/testing/core/dao"
	"github.com/mi991203/TestMoon/backend/testing/core/logic"
	"github.com/mi991203/TestMoon/backend/testing/core/model"
	"github.com/mi991203/TestMoon/backend/testing/core/serializer"
	p "github.com/mi991203/TestMoon/libs/go/platform"
	pb "github.com/mi991203/TestMoon/protocols/testing/core/v1"
)

func (s *Server) ListProjects(ctx context.Context, r *pb.ListProjectsRequest) (*pb.ListProjectsResponse, error) {
	if err := validateRequest(r); err != nil {
		return nil, err
	}
	deps := s.deps
	if e := p.RequireOrg(ctx, r.OrgId); e != nil {
		return nil, e
	}
	q := deps.DB.Where("org_id = ?", r.OrgId)
	if p.Actor(ctx).Organizations[r.OrgId] != "admin" {
		var members []model.Member
		if e := deps.DB.Where("org_id = ? AND user_id = ?", r.OrgId, p.Actor(ctx).UserID).Find(&members).Error; e != nil {
			return nil, p.DBError(e)
		}
		ids := []string{}
		for _, v := range members {
			ids = append(ids, v.ProjectID)
		}
		q = q.Where("id IN ?", ids)
	}
	var items []model.Project
	if e := q.Order("name,id").Find(&items).Error; e != nil {
		return nil, p.DBError(e)
	}
	return serializer.ListProjects(map[string]any{"items": items})
}

func (s *Server) CreateProject(ctx context.Context, r *pb.CreateProjectRequest) (*pb.CreateProjectResponse, error) {
	if err := validateRequest(r); err != nil {
		return nil, err
	}
	deps := s.deps
	var in struct {
		Name        string `json:"name" validate:"required,max=128"`
		Description string `json:"description" validate:"max=4096"`
	}
	if e := p.Decode(r, &in); e != nil {
		return nil, e
	}
	var project model.Project

	if e := p.RequireOrgAdmin(ctx, r.OrgId); e != nil {
		return nil, e
	}
	project = model.Project{Base: p.NewBase(), OrgID: r.OrgId, Name: in.Name, Description: in.Description}
	if e := deps.DB.Transaction(func(tx *gorm.DB) error {
		if e := tx.Create(&project).Error; e != nil {
			return e
		}
		return tx.Create(&model.Member{Base: p.NewBase(), OrgID: r.OrgId, ProjectID: project.ID, UserID: p.Actor(ctx).UserID, Role: "admin"}).Error
	}); e != nil {
		return nil, p.DBError(e)
	}

	return serializer.CreateProject(map[string]any{"project": project})
}

func (s *Server) UpdateProject(ctx context.Context, r *pb.UpdateProjectRequest) (*pb.UpdateProjectResponse, error) {
	if err := validateRequest(r); err != nil {
		return nil, err
	}
	deps := s.deps
	var in struct {
		Name        string `json:"name" validate:"required,max=128"`
		Description string `json:"description" validate:"max=4096"`
	}
	if e := p.Decode(r, &in); e != nil {
		return nil, e
	}
	var project model.Project

	var e error
	project, _, e = logic.RequireProject(ctx, deps.DB, r.OrgId, r.Id, "manage")
	if e != nil {
		return nil, e
	}
	if e = deps.DB.Model(&project).Updates(map[string]any{"name": in.Name, "description": in.Description}).Error; e != nil {
		return nil, p.DBError(e)
	}
	project.Name = in.Name
	project.Description = in.Description

	return serializer.UpdateProject(map[string]any{"project": project})
}

func (s *Server) DeleteProject(ctx context.Context, r *pb.DeleteProjectRequest) (*pb.DeleteProjectResponse, error) {
	if err := validateRequest(r); err != nil {
		return nil, err
	}
	deps := s.deps
	if _, _, e := logic.RequireProject(ctx, deps.DB, r.OrgId, r.ProjectId, "manage"); e != nil {
		return nil, e
	}
	var n int64
	if e := dao.WithProjectScope(deps.DB.Model(&model.Resource{}), r.OrgId, r.ProjectId).Count(&n).Error; e != nil {
		return nil, p.DBError(e)
	}
	if n > 0 {
		return nil, p.Conflict("请先删除项目资源；项目删除只允许空项目")
	}
	if e := logic.DeleteEmptyProject(ctx, deps, r.OrgId, r.ProjectId); e != nil {
		return nil, e
	}
	return serializer.DeleteProject(map[string]any{"ok": true})
}
