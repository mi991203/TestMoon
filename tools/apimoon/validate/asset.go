package validate

import (
	"encoding/json"
	"slices"
	"strings"
	"unicode/utf8"

	domain "github.com/mi991203/TestMoon/libs/go/testing/domain"
	"github.com/mi991203/TestMoon/tools/apimoon/clierror"
	"github.com/mi991203/TestMoon/tools/apimoon/jsonutil"
	"github.com/mi991203/TestMoon/tools/apimoon/platform"
)

func validateDocument(kind string, doc map[string]any) error {
	b, _ := json.Marshal(doc)
	if doc == nil || len(b) > 2<<20 {
		return clierror.New("validation", "缺少 document 或资源超过 2MiB")
	}
	switch kind {
	case "api":
		var v domain.API
		_ = json.Unmarshal(b, &v)
		if v.Protocol != "http" {
			return clierror.New("validation", "当前仅支持 http")
		}
		return validateRequest(v.Request)
	case "case":
		var v domain.Case
		_ = json.Unmarshal(b, &v)
		return validateCase(v)
	case "scenario":
		var v domain.Scenario
		_ = json.Unmarshal(b, &v)
		steps, err := domain.FlattenSteps(v.Steps)
		if err != nil {
			return clierror.New("validation", err.Error())
		}
		if len(steps) == 0 {
			return clierror.New("validation", "场景不能为空")
		}
		for _, step := range steps {
			if step.Inline != nil {
				if err = validateCase(*step.Inline); err != nil {
					return err
				}
			}
			if err = validatePatch(step.Override); err != nil {
				return err
			}
		}
	case "task":
		var v domain.Task
		_ = json.Unmarshal(b, &v)
		items, err := domain.FlattenSelections(v.Items, 100)
		if err != nil {
			return clierror.New("validation", err.Error())
		}
		if len(items) == 0 || v.TimeoutSeconds < 0 || v.TimeoutSeconds > 3600 {
			return clierror.New("validation", "任务不能为空，超时应为 0..3600 秒")
		}
	case "environment":
		var v domain.Environment
		_ = json.Unmarshal(b, &v)
		for k, x := range v.Variables {
			if k == "" || len(k) > 128 || len(x.Value) > 65536 {
				return clierror.New("validation", "环境变量超出平台限制")
			}
		}
	case "folder":
		if !slices.Contains([]string{"api", "case", "scenario"}, jsonutil.String(doc["kind"])) {
			return clierror.New("validation", "目录类型必须是 api/case/scenario")
		}
	default:
		return clierror.New("validation", "未知资源类型")
	}
	return nil
}
func Asset(kind, action string, input map[string]any) error {
	svc, name, err := platform.ResourceMethod(kind, action)
	if err != nil {
		return err
	}
	if kind == "task" && action == "get" {
		name = "DeleteTask"
	} // same scoped ID shape; task detail is read through ListTasks.
	canonical, err := platform.CanonicalRequest(svc, name, input)
	if err != nil {
		return err
	}
	var normalized map[string]any
	if err = jsonutil.Decode(canonical, &normalized); err != nil {
		return err
	}
	if action == "get" || action == "delete" || action == "update" {
		if jsonutil.String(input["id"]) == "" {
			return clierror.New("validation", "缺少资源 ID")
		}
	}
	if action == "create" || action == "update" {
		name := jsonutil.String(input["name"])
		if strings.TrimSpace(name) == "" || utf8.RuneCountInString(name) > 128 {
			return clierror.New("validation", "名称应为 1..128 个字符")
		}
		return validateDocument(kind, jsonutil.Object(normalized["document"]))
	}
	return nil
}
