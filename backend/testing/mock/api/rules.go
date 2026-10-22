package api

import (
	"context"
	"encoding/json"

	"github.com/mi991203/TestMoon/backend/testing/mock/model"
	"github.com/mi991203/TestMoon/backend/testing/mock/serializer"
	p "github.com/mi991203/TestMoon/libs/go/platform"
	pb "github.com/mi991203/TestMoon/protocols/testing/mock/v1"
)

func (s *Server) ListRules(ctx context.Context, r *pb.ListRulesRequest) (*pb.ListRulesResponse, error) {
	if err := validateRequest(r); err != nil {
		return nil, err
	}
	deps := s.deps
	role, authErr := deps.Core.Role(ctx, r.OrgId, r.ProjectId, "read")
	if authErr != nil {
		return nil, authErr
	}
	q := deps.DB.Where("org_id = ? AND project_id = ?", r.OrgId, r.ProjectId)
	var all []model.Rule
	if e := q.Order("priority DESC,id").Find(&all).Error; e != nil {
		return nil, p.DBError(e)
	}
	items := []any{}
	for _, v := range all {
		response, e := deps.Rules.Response(v)
		if e != nil {
			return nil, e
		}
		var doc any = response
		if role == "viewer" {
			doc = p.Sanitize(doc, nil)
		}
		item := map[string]any{"id": v.ID, "name": v.Name, "endpoint_id": v.EndpointID, "method": v.Method, "path": v.Path, "priority": v.Priority, "document": doc}
		items = append(items, item)
	}
	return serializer.ListRules(map[string]any{"items": items})
}

func (s *Server) CreateRule(ctx context.Context, r *pb.CreateRuleRequest) (*pb.CreateRuleResponse, error) {
	if err := validateRequest(r); err != nil {
		return nil, err
	}
	if r.Document == nil {
		return nil, p.Bad("缺少Mock响应配置")
	}
	deps := s.deps
	_, authErr := deps.Core.Role(ctx, r.OrgId, r.ProjectId, "write")
	if authErr != nil {
		return nil, authErr
	}
	in := ruleInput{}
	if e := p.Decode(r, &in); e != nil {
		return nil, e
	}
	var endpoint model.Endpoint
	if e := deps.DB.First(&endpoint, "org_id = ? AND project_id = ? AND id = ?", r.OrgId, r.ProjectId, in.EndpointID).Error; e != nil {
		return nil, p.DBError(e)
	}
	if err := validateRuleDefinition(in.Path, in.Document); err != nil {
		return nil, err
	}
	b, _ := json.Marshal(in.Document)
	v := model.Rule{Base: p.NewBase(), OrgID: r.OrgId, ProjectID: r.ProjectId, EndpointID: in.EndpointID, Name: in.Name, Method: in.Method, Path: in.Path, Priority: in.Priority, Document: p.Seal(deps.Config.EncryptionKey, string(b))}

	if e := deps.Rules.Create(deps.DB, &v); e != nil {
		return nil, p.DBError(e)
	}

	return serializer.CreateRule(map[string]any{"id": v.ID})
}

func (s *Server) UpdateRule(ctx context.Context, r *pb.UpdateRuleRequest) (*pb.UpdateRuleResponse, error) {
	if err := validateRequest(r); err != nil {
		return nil, err
	}
	if r.Document == nil {
		return nil, p.Bad("缺少Mock响应配置")
	}
	deps := s.deps
	_, authErr := deps.Core.Role(ctx, r.OrgId, r.ProjectId, "write")
	if authErr != nil {
		return nil, authErr
	}
	in := ruleInput{}
	if e := p.Decode(r, &in); e != nil {
		return nil, e
	}
	var endpoint model.Endpoint
	if e := deps.DB.First(&endpoint, "org_id = ? AND project_id = ? AND id = ?", r.OrgId, r.ProjectId, in.EndpointID).Error; e != nil {
		return nil, p.DBError(e)
	}
	if err := validateRuleDefinition(in.Path, in.Document); err != nil {
		return nil, err
	}
	b, _ := json.Marshal(in.Document)
	v := model.Rule{Base: p.NewBase(), OrgID: r.OrgId, ProjectID: r.ProjectId, EndpointID: in.EndpointID, Name: in.Name, Method: in.Method, Path: in.Path, Priority: in.Priority, Document: p.Seal(deps.Config.EncryptionKey, string(b))}

	v.ID = r.Id
	if err := deps.Rules.Update(deps.DB, v); err != nil {
		return nil, err
	}

	return serializer.UpdateRule(map[string]any{"id": v.ID})
}

func (s *Server) DeleteRule(ctx context.Context, r *pb.DeleteRuleRequest) (*pb.DeleteRuleResponse, error) {
	if err := validateRequest(r); err != nil {
		return nil, err
	}
	deps := s.deps
	_, authErr := deps.Core.Role(ctx, r.OrgId, r.ProjectId, "write")
	if authErr != nil {
		return nil, authErr
	}
	q := deps.DB.Where("org_id = ? AND project_id = ?", r.OrgId, r.ProjectId)
	result := q.Where("id = ?", r.Id).Delete(&model.Rule{})
	if result.Error != nil {
		return nil, p.DBError(result.Error)
	}
	return serializer.DeleteRule(map[string]any{"ok": true})
}
