package plan

import (
	"context"
	"strconv"
	"strings"

	"github.com/mi991203/TestMoon/tools/apimoon/clierror"
	"github.com/mi991203/TestMoon/tools/apimoon/jsonutil"
	"github.com/mi991203/TestMoon/tools/apimoon/platform"
	"github.com/mi991203/TestMoon/tools/apimoon/validate"
)

func Order(p Plan) ([]Entry, error) {
	if p.Version != 1 || strings.TrimSpace(p.PlanID) == "" || len(p.PlanID) > 128 || len(p.Entries) == 0 || len(p.Entries) > 2000 {
		return nil, clierror.New("validation", "计划需要 version=1、1..128 字符 plan_id 和 1..2000 个条目")
	}
	entries := map[string]Entry{}
	targets := map[string]bool{}
	for _, e := range p.Entries {
		if e.Key == "" || len(e.Key) > 128 || strings.HasPrefix(e.Key, "ref:") {
			return nil, clierror.New("validation", "条目 key 必须是 1..128 字符且不能以 ref: 开头")
		}
		if _, ok := entries[e.Key]; ok {
			return nil, clierror.New("validation", "条目 key 重复: "+e.Key)
		}
		if _, ok := platform.Types[e.Kind]; !ok {
			return nil, clierror.New("validation", "未知资源类型: "+e.Kind)
		}
		if e.ID != "" {
			rev, err := strconv.ParseInt(string(e.Revision), 10, 64)
			if err != nil || rev < 1 {
				return nil, clierror.New("validation", e.Key+": 更新必须提供正整数 revision")
			}
			if targets[e.ID] {
				return nil, clierror.New("validation", "同一平台 ID 不能绑定多个条目")
			}
			targets[e.ID] = true
		} else if e.Revision != "" {
			return nil, clierror.New("validation", e.Key+": revision 需要同时提供 id")
		}
		action := "create"
		if e.ID != "" {
			action = "update"
		}
		if err := validate.Asset(e.Kind, action, e.input()); err != nil {
			return nil, clierror.New("validation", e.Key+": "+err.Error())
		}
		entries[e.Key] = e
	}
	var ordered []Entry
	state := map[string]int{}
	var visit func(string) error
	visit = func(key string) error {
		if state[key] == 2 {
			return nil
		}
		if state[key] == 1 {
			return clierror.New("validation", "计划引用存在循环: "+key)
		}
		state[key] = 1
		e := entries[key]
		err := references(e, func(r Reference) (string, error) {
			if !strings.HasPrefix(r.ID, "ref:") {
				return r.ID, nil
			}
			ref := strings.TrimPrefix(r.ID, "ref:")
			target, ok := entries[ref]
			if !ok {
				return "", clierror.New("validation", key+"."+r.Path+": 引用不存在: "+ref)
			}
			if target.Kind != r.Kind {
				return "", clierror.New("validation", key+"."+r.Path+": 引用类型不匹配")
			}
			return r.ID, visit(ref)
		})
		if err != nil {
			return err
		}
		state[key] = 2
		ordered = append(ordered, e)
		return nil
	}
	for _, e := range p.Entries {
		if err := visit(e.Key); err != nil {
			return nil, err
		}
	}
	return ordered, nil
}
func Validate(ctx context.Context, c *platform.Client, p Plan) ([]Entry, error) {
	p.Entries = append([]Entry(nil), p.Entries...)
	for i := range p.Entries {
		p.Entries[i].Document = jsonutil.Clone(p.Entries[i].Document)
		if p.Entries[i].ID == "" {
			in := p.Entries[i].input()
			c.Defaults(p.Entries[i].Kind, in)
			p.Entries[i].Document = jsonutil.Object(in["document"])
		}
	}
	ordered, err := Order(p)
	if err != nil {
		return nil, err
	}
	if _, err = c.Call(ctx, "core", "Authorize", jsonutil.Merge(c.Scope(), map[string]any{"permission": "write"})); err != nil {
		return nil, err
	}
	checked := map[string]bool{}
	runners := map[string]bool{}
	for _, e := range ordered {
		if e.Kind == "task" {
			if _, err = c.Call(ctx, "core", "Authorize", jsonutil.Merge(c.Scope(), map[string]any{"permission": "execute"})); err != nil {
				return nil, err
			}
		}
		if e.ID != "" {
			v, er := c.Resource(ctx, e.Kind, "get", map[string]any{"id": e.ID})
			if er != nil {
				return nil, er
			}
			if jsonutil.String(jsonutil.Object(v["resource"])["revision"]) != string(e.Revision) {
				return nil, clierror.New("conflict", e.Key+": 平台版本已改变")
			}
		}
		err = references(e, func(r Reference) (string, error) {
			if strings.HasPrefix(r.ID, "ref:") {
				return r.ID, nil
			}
			key := r.Kind + ":" + r.ID
			if !checked[key] {
				if _, er := c.Resource(ctx, r.Kind, "get", map[string]any{"id": r.ID}); er != nil {
					return "", clierror.New("validation", e.Key+"."+r.Path+": 目标不可访问 ("+clierror.From(er).Code+")")
				}
				checked[key] = true
			}
			return r.ID, nil
		})
		if err != nil {
			return nil, err
		}
		in := e.input()
		if runner := jsonutil.String(jsonutil.Object(in["document"])["runner_id"]); runner != "" && !runners[runner] {
			v, er := c.Call(ctx, "execution", "ListProjectRunners", c.Scope())
			if er != nil {
				return nil, er
			}
			found := false
			for _, it := range v["items"].([]any) {
				if jsonutil.String(jsonutil.Object(it)["id"]) == runner {
					found = true
				}
			}
			if !found {
				return nil, clierror.New("validation", e.Key+": Runner 不属于当前项目或已撤销")
			}
			runners[runner] = true
		}
	}
	return ordered, nil
}
