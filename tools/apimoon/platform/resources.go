package platform

import (
	"context"
	"net/url"
	"strings"

	"github.com/mi991203/TestMoon/tools/apimoon/clierror"
	"github.com/mi991203/TestMoon/tools/apimoon/jsonutil"
)

type resourceType struct{ Singular, Plural, Service string }

var Types = map[string]resourceType{
	"api": {"API", "APIs", "core"}, "case": {"TestCase", "TestCases", "core"}, "scenario": {"Scenario", "Scenarios", "core"},
	"environment": {"Environment", "Environments", "core"}, "folder": {"Folder", "Folders", "core"}, "task": {"Task", "Tasks", "execution"},
}

func (c *Client) Scope() map[string]any {
	return map[string]any{"org_id": c.Config.Context.OrgID, "project_id": c.Config.Context.ProjectID}
}
func (c *Client) Scoped(input map[string]any) (map[string]any, error) {
	out := jsonutil.Clone(input)
	for k, v := range c.Scope() {
		if v == "" {
			return nil, clierror.New("config", "缺少 "+k)
		}
		if got, ok := out[k]; ok && got != v {
			return nil, clierror.New("validation", k+" 与当前作用域不一致")
		}
		out[k] = v
	}
	return out, nil
}
func ResourceMethod(kind, action string) (string, string, error) {
	r, ok := Types[kind]
	if !ok {
		return "", "", clierror.New("validation", "未知资源类型: "+kind)
	}
	if action == "list" {
		return r.Service, "List" + r.Plural, nil
	}
	verbs := map[string]string{"get": "Get", "create": "Create", "update": "Update", "delete": "Delete"}
	v, ok := verbs[action]
	if !ok {
		return "", "", clierror.New("validation", "未知资源操作")
	}
	return r.Service, v + r.Singular, nil
}
func (c *Client) Resource(ctx context.Context, kind, action string, input map[string]any) (map[string]any, error) {
	service, name, err := ResourceMethod(kind, action)
	if err != nil {
		return nil, err
	}
	in, err := c.Scoped(input)
	if err != nil {
		return nil, err
	}
	if action == "update" && (jsonutil.String(in["revision"]) == "" || jsonutil.String(in["revision"]) == "0") {
		return nil, clierror.New("validation", "更新必须提供 revision")
	}
	if kind == "task" && action == "get" {
		v, e := c.Call(ctx, "execution", "ListTasks", c.Scope())
		if e != nil {
			return nil, e
		}
		for _, it := range v["items"].([]any) {
			if jsonutil.Object(it)["id"] == in["id"] {
				return map[string]any{"resource": it}, nil
			}
		}
		return nil, &clierror.Error{Code: "platform", Message: "任务不存在", Status: 404}
	}
	v, err := c.Call(ctx, service, name, in)
	if err != nil {
		return nil, err
	}
	if kind == "task" && (action == "create" || action == "update") {
		id := jsonutil.String(v["id"])
		if id == "" {
			return nil, clierror.New("protocol", "任务响应缺少 ID")
		}
		resource, readErr := c.Resource(ctx, kind, "get", map[string]any{"id": id})
		if readErr != nil {
			return nil, &clierror.Error{Code: "uncertain", Message: "任务已写入，但无法读取确认版本", Details: map[string]any{"id": id, "cause": clierror.From(readErr).Code}}
		}
		return resource, nil
	}
	return v, nil
}
func (c *Client) Defaults(kind string, input map[string]any) {
	doc := jsonutil.Object(input["document"])
	if doc == nil {
		return
	}
	if kind == "api" || kind == "case" || kind == "task" {
		if _, ok := doc["environment_id"]; !ok && c.Config.Context.EnvironmentID != "" {
			doc["environment_id"] = c.Config.Context.EnvironmentID
		}
		if _, ok := doc["runner_id"]; !ok && c.Config.Context.RunnerID != "" {
			doc["runner_id"] = c.Config.Context.RunnerID
		}
	}
}
func (c *Client) Link(kind, id string) string {
	route := map[string]string{"api": "api-testing/apis/" + id + "/debug", "case": "api-testing/cases/" + id, "scenario": "scenario/" + id, "environment": "environment/" + id, "task": "tasks/" + id, "run": "reports/" + id}[kind]
	if route == "" {
		return ""
	}
	q := url.Values{"org": {c.Config.Context.OrgID}, "project": {c.Config.Context.ProjectID}}
	return strings.TrimRight(c.Config.BaseURL, "/") + "/apps/testing/" + route + "?" + q.Encode()
}
