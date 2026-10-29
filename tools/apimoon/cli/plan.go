package cli

import (
	"context"
	"io"

	"github.com/mi991203/TestMoon/tools/apimoon/clierror"
	"github.com/mi991203/TestMoon/tools/apimoon/plan"
	"github.com/mi991203/TestMoon/tools/apimoon/platform"
)

func planCommand(ctx context.Context, c *platform.Client, file string, o options, action string, stdin io.Reader) (any, error) {
	var p plan.Plan
	if err := readInput(o.file, stdin, &p); err != nil {
		return nil, err
	}
	if _, err := plan.Order(p); err != nil {
		return nil, err
	}
	s := plan.Syncer{Client: c, ConfigFile: file}
	switch action {
	case "validate":
		entries, err := plan.Validate(ctx, c, p)
		if err != nil {
			return nil, err
		}
		keys := []string{}
		for _, e := range entries {
			keys = append(keys, e.Key)
		}
		return map[string]any{"plan_id": p.PlanID, "order": keys, "warnings": []string{"静态预校验不执行脚本或请求；动态变量、断言正确性及部署配置需运行验证"}}, nil
	case "status":
		return s.Status(p.PlanID)
	case "sync":
		state, err := s.Sync(ctx, p)
		links := map[string]string{}
		for k, b := range state.Entries {
			if b.State == "confirmed" {
				links[k] = c.Link(b.Kind, b.ID)
			}
		}
		return map[string]any{"state": state, "urls": links}, err
	case "bind":
		if o.key == "" || o.id == "" {
			return nil, clierror.New("usage", "plan bind 需要 --key 和 --id")
		}
		return s.Bind(ctx, p, o.key, o.id)
	default:
		return nil, clierror.New("usage", "未知 plan 子命令")
	}
}
