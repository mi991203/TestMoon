package plan

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/mi991203/TestMoon/tools/apimoon/clierror"
	"github.com/mi991203/TestMoon/tools/apimoon/jsonutil"
	"github.com/mi991203/TestMoon/tools/apimoon/validate"
)

func (s Syncer) resolve(e Entry, state PlanState) (map[string]any, error) {
	e.Document = jsonutil.Clone(e.Document)
	resolve := func(r Reference) (string, error) {
		if !strings.HasPrefix(r.ID, "ref:") {
			return r.ID, nil
		}
		b, ok := state.Entries[strings.TrimPrefix(r.ID, "ref:")]
		if !ok || b.State != "confirmed" || b.ID == "" {
			return "", clierror.New("conflict", "引用尚未完成同步: "+r.ID)
		}
		return b.ID, nil
	}
	if err := references(e, resolve); err != nil {
		return nil, err
	}
	if e.FolderID != "" {
		v, err := resolve(Reference{ID: e.FolderID})
		if err != nil {
			return nil, err
		}
		e.FolderID = v
	}
	return e.input(), nil
}
func (s Syncer) Sync(ctx context.Context, p Plan) (PlanState, error) {
	if s.Client.Config.Auth.UserID == "" {
		return PlanState{}, clierror.New("auth", "计划同步需要登录用户标识，请重新 auth login")
	}
	unlock, err := s.lock(p.PlanID)
	if err != nil {
		return PlanState{}, err
	}
	defer unlock()
	state, err := s.Status(p.PlanID)
	if err != nil {
		return state, err
	}
	for key, b := range state.Entries {
		if b.State != "confirmed" {
			return state, clierror.New("uncertain", key+": 存在结果不明的写入，请核对后使用 plan bind 恢复")
		}
	}
	// Known bindings supply update identity but never silently replace an explicit revision.
	check := p
	check.Entries = append([]Entry(nil), p.Entries...)
	for i, e := range check.Entries {
		check.Entries[i].Document = jsonutil.Clone(e.Document)
		if e.ID == "" {
			in := check.Entries[i].input()
			s.Client.Defaults(e.Kind, in)
			check.Entries[i].Document = jsonutil.Object(in["document"])
		}
		if b, ok := state.Entries[e.Key]; ok {
			if b.Kind != e.Kind || (e.ID != "" && e.ID != b.ID) {
				return state, clierror.New("conflict", e.Key+": 已绑定其他类型或 ID")
			}
			if e.ID == "" {
				check.Entries[i].ID = b.ID
				check.Entries[i].Revision = json.Number(b.Revision)
			}
		}
	}
	ordered, err := Validate(ctx, s.Client, check)
	if err != nil {
		return state, err
	}
	for _, e := range ordered {
		in, er := s.resolve(e, state)
		if er != nil {
			return state, er
		}
		action := "create"
		if e.ID != "" {
			action = "update"
		} else {
			s.Client.Defaults(e.Kind, in)
		}
		hash := assetHash(e.Kind, in)
		b, known := state.Entries[e.Key]
		if known && b.Hash == hash {
			continue
		}
		if er = validate.Asset(e.Kind, action, in); er != nil {
			return state, er
		}
		pending := Binding{Kind: e.Kind, ID: e.ID, Revision: string(e.Revision), Hash: b.Hash, PendingHash: hash, State: "pending"}
		state.Entries[e.Key] = pending
		if er = s.save(&state); er != nil {
			return state, er
		}
		v, er := s.Client.Resource(ctx, e.Kind, action, in)
		if er != nil {
			cause := clierror.From(er)
			if cause.Status >= 400 && cause.Status < 500 {
				if known {
					state.Entries[e.Key] = b
				} else {
					delete(state.Entries, e.Key)
				}
				if saveErr := s.save(&state); saveErr != nil {
					return state, saveErr
				}
				return state, er
			}
			return state, &clierror.Error{Code: "uncertain", Message: e.Key + ": 写入结果不明，禁止自动重试，请核对平台后运行 plan bind", Details: map[string]any{"cause": cause.Code, "key": e.Key}}
		}
		r := jsonutil.Object(v["resource"])
		if jsonutil.String(r["id"]) == "" || jsonutil.String(r["revision"]) == "" {
			return state, clierror.New("uncertain", e.Key+": 保存响应缺少资产标识，请使用 plan bind")
		}
		state.Entries[e.Key] = Binding{Kind: e.Kind, ID: jsonutil.String(r["id"]), Revision: jsonutil.String(r["revision"]), Hash: hash, State: "confirmed"}
		if er = s.save(&state); er != nil {
			return state, er
		}
	}
	return state, nil
}
func (s Syncer) Bind(ctx context.Context, p Plan, key, id string) (PlanState, error) {
	unlock, err := s.lock(p.PlanID)
	if err != nil {
		return PlanState{}, err
	}
	defer unlock()
	state, err := s.Status(p.PlanID)
	if err != nil {
		return state, err
	}
	b, ok := state.Entries[key]
	if !ok || b.State == "confirmed" {
		return state, clierror.New("conflict", "bind 仅用于恢复待确认写入")
	}
	if b.ID != "" && b.ID != id {
		return state, clierror.New("conflict", "更新写入只能绑定原平台 ID")
	}
	var entry *Entry
	for _, e := range p.Entries {
		if e.Key == key {
			copy := e
			entry = &copy
			break
		}
	}
	if entry == nil || entry.Kind != b.Kind {
		return state, clierror.New("validation", "计划中缺少对应条目")
	}
	for other, existing := range state.Entries {
		if other != key && existing.ID == id {
			return state, clierror.New("conflict", "平台 ID 已绑定其他条目")
		}
	}
	in, err := s.resolve(*entry, state)
	if err != nil {
		return state, err
	}
	if entry.ID == "" {
		s.Client.Defaults(entry.Kind, in)
	}
	if assetHash(entry.Kind, in) != b.PendingHash {
		return state, clierror.New("conflict", "计划或默认配置已变化，需使用写入时的计划与配置核对")
	}
	v, err := s.Client.Resource(ctx, entry.Kind, "get", map[string]any{"id": id})
	if err != nil {
		return state, err
	}
	r := jsonutil.Object(v["resource"])
	actual := map[string]any{"name": r["name"], "document": r["document"]}
	if entry.Kind != "task" {
		if folder := jsonutil.String(r["folder_id"]); folder != "" {
			actual["folder_id"] = folder
		}
	}
	if assetHash(entry.Kind, actual) != b.PendingHash {
		return state, clierror.New("conflict", "平台内容与待确认写入不一致，拒绝绑定")
	}
	state.Entries[key] = Binding{Kind: entry.Kind, ID: id, Revision: jsonutil.String(r["revision"]), Hash: b.PendingHash, State: "confirmed"}
	return state, s.save(&state)
}
