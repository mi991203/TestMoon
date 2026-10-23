package logic

import (
	"context"
	"encoding/json"

	"gorm.io/gorm"

	"github.com/mi991203/TestMoon/backend/testing/core/internal/runtime"
	"github.com/mi991203/TestMoon/backend/testing/core/internal/validation"
	p "github.com/mi991203/TestMoon/libs/go/platform"
	m "github.com/mi991203/TestMoon/libs/go/testing/domain"
)

type SnapshotInput struct {
	OrgID     string `json:"org_id"`
	ProjectID string `json:"project_id"`
	m.Task
	Debug *m.Case `json:"debug"`
}

func BuildSnapshot(ctx context.Context, deps *runtime.Dependencies, in SnapshotInput) (m.Snapshot, error) {
	selections, err := m.FlattenSelections(in.Items, deps.Config.MaxItems)
	if err != nil {
		return m.Snapshot{}, p.Bad(err.Error())
	}
	out := m.Snapshot{Version: 1, OrgID: in.OrgID, ProjectID: in.ProjectID, Revisions: map[string]int64{}, Environment: m.Environment{Variables: map[string]m.Variable{}}, TimeoutSeconds: in.TimeoutSeconds}
	if out.TimeoutSeconds == 0 {
		out.TimeoutSeconds = 900
	}
	if out.TimeoutSeconds < 1 || out.TimeoutSeconds > deps.Config.MaxRunSeconds {
		return m.Snapshot{}, p.Bad("执行超时超出配置范围")
	}
	e := deps.DB.Transaction(func(tx *gorm.DB) error {
		if in.EnvironmentID != "" {
			a, e := LoadResource(deps, tx, in.OrgID, in.ProjectID, in.EnvironmentID, "environment")
			if e != nil {
				return e
			}
			env, e := DecodeDocument[m.Environment](deps, a)
			if e != nil {
				return e
			}
			out.Environment = env
			out.Revisions[a.ID] = a.Revision
		}
		compileCase := func(c m.Case) (m.Case, error) {
			if c.APIID != "" {
				a, e := LoadResource(deps, tx, in.OrgID, in.ProjectID, c.APIID, "api")
				if e != nil {
					return c, e
				}
				api, e := DecodeDocument[m.API](deps, a)
				if e != nil {
					return c, e
				}
				if api.Protocol != "http" {
					return c, p.Bad("不支持该协议")
				}
				c.Request = api.Request
				out.Revisions[a.ID] = a.Revision
			}
			c.Request = m.Merge(c.Request, c.Override)
			if e := validation.Request(c.Request); e != nil {
				return c, e
			}
			if e := validateOperations(c.Before); e != nil {
				return c, e
			}
			if e := validateOperations(c.After); e != nil {
				return c, e
			}
			return c, nil
		}
		if in.Debug != nil {
			c, e := compileCase(*in.Debug)
			if e != nil {
				return e
			}
			name := c.Name
			if name == "" {
				name = "接口调试"
			}
			out.Items = append(out.Items, m.CompiledItem{ID: p.NewBase().ID, Name: name, CollectionVariables: c.CollectionVariables, Steps: []m.CompiledStep{{ID: p.NewBase().ID, Name: name, Case: c}}})
		}
		for _, sel := range selections {
			if sel.Kind != "case" && sel.Kind != "scenario" {
				return p.Bad("任务只支持case或scenario")
			}
			a, e := LoadResource(deps, tx, in.OrgID, in.ProjectID, sel.ID, sel.Kind)
			if e != nil {
				return e
			}
			out.Revisions[a.ID] = a.Revision
			item := m.CompiledItem{ID: a.ID, Name: a.Name}
			if sel.NodeID != "" {
				item.ID = sel.NodeID
			}
			if sel.Kind == "case" {
				c, e := DecodeDocument[m.Case](deps, a)
				if e != nil {
					return e
				}
				c, e = compileCase(c)
				if e != nil {
					return e
				}
				item.Steps = []m.CompiledStep{{ID: a.ID, Name: a.Name, Case: c}}
				item.CollectionVariables = c.CollectionVariables
			} else {
				sc, e := DecodeDocument[m.Scenario](deps, a)
				if e != nil {
					return e
				}
				steps, err := m.FlattenSteps(sc.Steps)
				if err != nil {
					return p.Bad(err.Error())
				}
				if len(steps) == 0 {
					return p.Bad("场景步骤数量超出限制")
				}
				item.CollectionVariables = sc.CollectionVariables
				for _, step := range steps {
					var c m.Case
					if step.CaseID != "" {
						a, e := LoadResource(deps, tx, in.OrgID, in.ProjectID, step.CaseID, "case")
						if e != nil {
							return e
						}
						c, e = DecodeDocument[m.Case](deps, a)
						if e != nil {
							return e
						}
						out.Revisions[a.ID] = a.Revision
						if step.Name == "" {
							step.Name = a.Name
						}
					} else if step.Inline != nil {
						c = *step.Inline
					} else {
						return p.Bad("场景步骤缺少用例或请求")
					}
					c, e = compileCase(c)
					if e != nil {
						return e
					}
					c.Request = m.Merge(c.Request, step.Override)
					vars := map[string]string{}
					for k, v := range c.Variables {
						vars[k] = v
					}
					for k, v := range step.Variables {
						vars[k] = v
					}
					c.Variables = vars
					if step.ID == "" {
						step.ID = p.NewBase().ID
					}
					item.Steps = append(item.Steps, m.CompiledStep{ID: step.ID, Name: step.Name, Case: c, ContinueOnFailure: step.ContinueOnFailure})
				}
			}
			out.Items = append(out.Items, item)
		}
		if len(out.Items) == 0 {
			return p.Bad("请选择用例或场景")
		}
		steps := 0
		for _, item := range out.Items {
			steps += len(item.Steps)
		}
		if steps > deps.Config.MaxSteps {
			return p.Bad("执行步骤总数超过配置限制")
		}
		encoded, _ := json.Marshal(out)
		if len(encoded) > 8<<20 {
			return p.Bad("执行快照超过8MiB")
		}
		return nil
	})
	if e != nil {
		return m.Snapshot{}, CoreError(e)
	}
	return out, nil
}
