package logic

import (
	"context"
	"encoding/json"

	"google.golang.org/grpc/status"
	"gorm.io/gorm"

	"github.com/mi991203/TestMoon/backend/testing/core/dao"
	"github.com/mi991203/TestMoon/backend/testing/core/internal/runtime"
	"github.com/mi991203/TestMoon/backend/testing/core/internal/validation"
	"github.com/mi991203/TestMoon/backend/testing/core/model"
	p "github.com/mi991203/TestMoon/libs/go/platform"
	m "github.com/mi991203/TestMoon/libs/go/testing/domain"
)

func SaveResource(ctx context.Context, deps *runtime.Dependencies, r *ResourceInput) (map[string]any, error) {
	if _, _, e := RequireProject(ctx, deps.DB, r.OrgId, r.ProjectId, "write"); e != nil {
		return nil, e
	}
	in := r
	var a model.Resource
	e := deps.DB.Transaction(func(tx *gorm.DB) error {
		if e := validateDocument(deps, tx, r, in.Kind, in.Document); e != nil {
			return e
		}
		if in.FolderID != "" {
			folder, e := LoadResource(deps, tx, r.OrgId, r.ProjectId, in.FolderID, "folder")
			if e != nil {
				return e
			}
			fd, e := readResourceDocument(deps, folder)
			if e != nil {
				return e
			}
			want := in.Kind
			if in.Kind == "folder" {
				want, _ = in.Document["kind"].(string)
			}
			if fd["kind"] != want {
				return p.Bad("目录类型不匹配")
			}
			parent := in.FolderID
			for depth := 0; parent != ""; depth++ {
				if parent == r.Id || depth >= 32 {
					return p.Bad("目录不能循环或超过32层")
				}
				f, e := LoadResource(deps, tx, r.OrgId, r.ProjectId, parent, "folder")
				if e != nil {
					return e
				}
				parent = f.FolderID
			}
		}
		b, _ := json.Marshal(in.Document)
		if len(b) > 2<<20 {
			return p.Bad("资源超过2MiB")
		}
		doc := p.Seal(deps.Config.EncryptionKey, string(b))
		if r.Id == "" {
			a = model.Resource{Base: p.NewBase(), OrgID: r.OrgId, ProjectID: r.ProjectId, Kind: in.Kind, Name: in.Name, FolderID: in.FolderID, Revision: 1, Document: doc}
			return deps.Assets.Create(tx, &a)
		}
		var e error
		a, e = LoadResource(deps, tx, r.OrgId, r.ProjectId, r.Id, in.Kind)
		if e != nil {
			return e
		}
		a.Name = in.Name
		a.FolderID = in.FolderID
		a.Document = doc
		return deps.Assets.Update(tx, &a, in.Revision)
	})
	if e != nil {
		return nil, CoreError(e)
	}
	item, e := resourceView(deps, a, "editor")
	if e != nil {
		return nil, e
	}
	return item, nil
}

func DecodeDocument[T any](deps *runtime.Dependencies, a model.Resource) (T, error) {
	var out T
	doc, e := readResourceDocument(deps, a)
	if e != nil {
		return out, e
	}
	b, _ := json.Marshal(doc)
	if e = json.Unmarshal(b, &out); e != nil {
		return out, p.Bad("文档格式无效")
	}
	return out, nil
}

func validateOperations(ops []m.Operation) error {
	if len(ops) > 20 {
		return p.Bad("前后置操作最多20项")
	}
	for _, v := range ops {
		if v.Type != "javascript" || v.Version != 1 || len(v.Script) > 65536 {
			return p.Bad("仅支持 version=1 的 javascript 操作，脚本最多64KiB")
		}
	}
	return nil
}

func validateDocument(deps *runtime.Dependencies, db *gorm.DB, r *ResourceInput, kind string, doc map[string]any) error {
	b, _ := json.Marshal(doc)
	switch kind {
	case "api":
		var v m.API
		if json.Unmarshal(b, &v) != nil {
			return p.Bad("接口格式无效")
		}
		if v.Protocol != "http" {
			return p.Bad("首版仅支持http")
		}
		return validation.Request(v.Request)
	case "case":
		var v m.Case
		if json.Unmarshal(b, &v) != nil {
			return p.Bad("用例格式无效")
		}
		if v.APIID != "" {
			if _, e := LoadResource(deps, db, r.OrgId, r.ProjectId, v.APIID, "api"); e != nil {
				return e
			}
		} else {
			if e := validation.Request(v.Request); e != nil {
				return e
			}
		}
		if e := validateOperations(v.Before); e != nil {
			return e
		}
		return validateOperations(v.After)
	case "scenario":
		var v m.Scenario
		if json.Unmarshal(b, &v) != nil || len(v.Steps) == 0 || len(v.Steps) > 100 {
			return p.Bad("场景应包含1..100个步骤")
		}
		for _, st := range v.Steps {
			if (st.CaseID == "") == (st.Inline == nil) {
				return p.Bad("每个步骤应引用用例或包含独立请求")
			}
			if st.CaseID != "" {
				if _, e := LoadResource(deps, db, r.OrgId, r.ProjectId, st.CaseID, "case"); e != nil {
					return e
				}
			} else {
				raw, _ := json.Marshal(st.Inline)
				d := map[string]any{}
				_ = json.Unmarshal(raw, &d)
				if e := validateDocument(deps, db, r, "case", d); e != nil {
					return e
				}
			}
		}
		return nil
	case "environment":
		var v m.Environment
		if json.Unmarshal(b, &v) != nil {
			return p.Bad("环境格式无效")
		}
		for k, x := range v.Variables {
			if k == "" || len(k) > 128 || len(x.Value) > 65536 {
				return p.Bad("环境变量名称或值超出限制")
			}
		}
		return nil
	case "folder":
		k, _ := doc["kind"].(string)
		if k != "api" && k != "case" && k != "scenario" {
			return p.Bad("目录仅支持api/case/scenario")
		}
		return nil
	}
	return p.Bad("不支持的资源类型")
}

func DeleteResource(ctx context.Context, deps *runtime.Dependencies, r *ResourceInput) error {
	if _, _, e := RequireProject(ctx, deps.DB, r.OrgId, r.ProjectId, "write"); e != nil {
		return e
	}
	e := deps.DB.Transaction(func(tx *gorm.DB) error {
		a, e := LoadResource(deps, tx, r.OrgId, r.ProjectId, r.Id, r.Kind)
		if e != nil {
			return e
		}
		if a.Kind == "folder" {
			var n int64
			dao.WithProjectScope(tx.Model(&model.Resource{}), r.OrgId, r.ProjectId).Where("folder_id = ?", a.ID).Count(&n)
			if n > 0 {
				return p.Conflict("目录不为空")
			}
		}
		var all []model.Resource
		if e := dao.WithProjectScope(tx, r.OrgId, r.ProjectId).Where("kind IN ?", []string{"case", "scenario"}).Find(&all).Error; e != nil {
			return e
		}
		for _, v := range all {
			if err := deps.Assets.Hydrate(tx, &v); err != nil {
				return err
			}
			if v.ID == a.ID {
				continue
			}
			if v.Kind == "case" {
				c, e := DecodeDocument[m.Case](deps, v)
				if e != nil {
					return e
				}
				if c.APIID == a.ID {
					return p.Conflict("接口仍被用例引用")
				}
			} else {
				sc, e := DecodeDocument[m.Scenario](deps, v)
				if e != nil {
					return e
				}
				for _, st := range sc.Steps {
					if st.CaseID == a.ID || (st.Inline != nil && st.Inline.APIID == a.ID) {
						return p.Conflict("资源仍被场景引用")
					}
				}
			}
		}
		return deps.Assets.Delete(tx, a)
	})
	if e != nil {
		return CoreError(e)
	}
	return nil
}

func CoreError(e error) error {
	if _, ok := status.FromError(e); ok {
		return e
	}
	return p.DBError(e)
}
