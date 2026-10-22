package api

import (
	"context"

	"gorm.io/gorm"

	"github.com/mi991203/TestMoon/backend/testing/mock/model"
	"github.com/mi991203/TestMoon/backend/testing/mock/serializer"
	p "github.com/mi991203/TestMoon/libs/go/platform"
	pb "github.com/mi991203/TestMoon/protocols/testing/mock/v1"
)

func (s *Server) ListEndpoints(ctx context.Context, r *pb.ListEndpointsRequest) (*pb.ListEndpointsResponse, error) {
	if err := validateRequest(r); err != nil {
		return nil, err
	}
	deps := s.deps
	_, authErr := deps.Core.Role(ctx, r.OrgId, r.ProjectId, "read")
	if authErr != nil {
		return nil, authErr
	}
	q := deps.DB.Where("org_id = ? AND project_id = ?", r.OrgId, r.ProjectId)
	var all []model.Endpoint
	if e := q.Order("name,id").Find(&all).Error; e != nil {
		return nil, p.DBError(e)
	}
	return serializer.ListEndpoints(map[string]any{"items": all})
}

func (s *Server) CreateEndpoint(ctx context.Context, r *pb.CreateEndpointRequest) (*pb.CreateEndpointResponse, error) {
	if err := validateRequest(r); err != nil {
		return nil, err
	}
	deps := s.deps
	_, authErr := deps.Core.Role(ctx, r.OrgId, r.ProjectId, "write")
	if authErr != nil {
		return nil, authErr
	}
	var in struct {
		Name string `json:"name" validate:"required,max=128"`
	}
	if e := p.Decode(r, &in); e != nil {
		return nil, e
	}
	v := model.Endpoint{Base: p.NewBase(), OrgID: r.OrgId, ProjectID: r.ProjectId, Name: in.Name}

	if e := deps.DB.Create(&v).Error; e != nil {
		return nil, p.DBError(e)
	}

	return serializer.CreateEndpoint(map[string]any{"endpoint": v, "path": "/m/" + v.ID})
}

func (s *Server) UpdateEndpoint(ctx context.Context, r *pb.UpdateEndpointRequest) (*pb.UpdateEndpointResponse, error) {
	if err := validateRequest(r); err != nil {
		return nil, err
	}
	deps := s.deps
	_, authErr := deps.Core.Role(ctx, r.OrgId, r.ProjectId, "write")
	if authErr != nil {
		return nil, authErr
	}
	q := deps.DB.Where("org_id = ? AND project_id = ?", r.OrgId, r.ProjectId)
	var in struct {
		Name string `json:"name" validate:"required,max=128"`
	}
	if e := p.Decode(r, &in); e != nil {
		return nil, e
	}
	v := model.Endpoint{Base: p.NewBase(), OrgID: r.OrgId, ProjectID: r.ProjectId, Name: in.Name}

	result := q.Model(&model.Endpoint{}).Where("id = ?", r.Id).Update("name", in.Name)
	if result.Error != nil {
		return nil, p.DBError(result.Error)
	}
	if result.RowsAffected == 0 {
		return nil, p.NotFound()
	}
	v.ID = r.Id

	return serializer.UpdateEndpoint(map[string]any{"endpoint": v, "path": "/m/" + v.ID})
}

func (s *Server) DeleteEndpoint(ctx context.Context, r *pb.DeleteEndpointRequest) (*pb.DeleteEndpointResponse, error) {
	if err := validateRequest(r); err != nil {
		return nil, err
	}
	deps := s.deps
	_, authErr := deps.Core.Role(ctx, r.OrgId, r.ProjectId, "write")
	if authErr != nil {
		return nil, authErr
	}
	e := deps.DB.Transaction(func(tx *gorm.DB) error {
		if e := tx.Where("org_id = ? AND project_id = ? AND endpoint_id = ?", r.OrgId, r.ProjectId, r.Id).Delete(&model.Rule{}).Error; e != nil {
			return e
		}
		return tx.Where("org_id = ? AND project_id = ? AND id = ?", r.OrgId, r.ProjectId, r.Id).Delete(&model.Endpoint{}).Error
	})
	if e != nil {
		return nil, p.DBError(e)
	}
	return serializer.DeleteEndpoint(map[string]any{"ok": true})
}
