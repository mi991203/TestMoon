package cli

import (
	"context"
	"fmt"
	"io"

	"github.com/google/uuid"
	"github.com/mi991203/TestMoon/tools/apimoon/clierror"
	"github.com/mi991203/TestMoon/tools/apimoon/execution"
	"github.com/mi991203/TestMoon/tools/apimoon/jsonutil"
	"github.com/mi991203/TestMoon/tools/apimoon/platform"
)

func runCommand(ctx context.Context, c *platform.Client, o options, group, action string, stderr io.Writer) (any, error) {
	if o.waitTimeout <= 0 || o.poll <= 0 {
		return nil, clierror.New("usage", "等待超时和轮询间隔必须大于 0")
	}
	if o.set["environment"] || o.set["runner"] {
		return nil, clierror.New("usage", "执行已有任务使用任务保存的环境和 Runner，请先显式更新任务")
	}
	if group == "run" && action == "list" {
		return c.Call(ctx, "execution", "ListRuns", c.Scope())
	}
	if o.id == "" {
		return nil, clierror.New("usage", "需要 --id")
	}
	if group == "task" {
		key := o.idempotency
		if key == "" {
			key = uuid.NewString()
		}
		v, err := c.Call(ctx, "execution", "RunTask", jsonutil.Merge(c.Scope(), map[string]any{"task_id": o.id, "idempotency_key": key}))
		if err != nil {
			return map[string]any{"idempotency_key": key, "task_id": o.id}, err
		}
		v["idempotency_key"] = key
		id := jsonutil.String(jsonutil.Object(v["run"])["id"])
		v["url"] = c.Link("run", id)
		if !o.wait {
			return v, nil
		}
		fmt.Fprintf(stderr, "已提交执行 %s\n", id)
		result, err := execution.Wait(ctx, c, id, o.waitTimeout, o.poll)
		if result == nil {
			result = v
		}
		result["idempotency_key"] = key
		return result, err
	}
	switch action {
	case "get":
		v, err := c.Call(ctx, "execution", "GetRun", jsonutil.Merge(c.Scope(), map[string]any{"id": o.id}))
		if err == nil {
			execution.Decorate(c, v)
		}
		return v, err
	case "cancel":
		return c.Call(ctx, "execution", "CancelRun", jsonutil.Merge(c.Scope(), map[string]any{"id": o.id}))
	case "wait":
		return execution.Wait(ctx, c, o.id, o.waitTimeout, o.poll)
	default:
		return nil, clierror.New("usage", "未知 run 子命令")
	}
}
