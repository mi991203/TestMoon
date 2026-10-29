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
	"github.com/spf13/cobra"
)

func (a *application) runCommand() *cobra.Command {
	group := commandGroup("run", "查询执行记录和报告，等待或取消执行")
	list := a.platformCommand("list", "列出当前项目的执行记录", projectScope, func(cmd *cobra.Command, _ []string, c *platform.Client, _ string) (any, error) {
		return c.Call(cmd.Context(), "execution", "ListRuns", c.Scope())
	})
	get := a.platformCommand("get", "读取执行报告；查询成功不代表测试通过", projectScope, func(cmd *cobra.Command, _ []string, c *platform.Client, _ string) (any, error) {
		v, err := c.Call(cmd.Context(), "execution", "GetRun", jsonutil.Merge(c.Scope(), map[string]any{"id": a.options.id}))
		if err == nil {
			execution.Decorate(c, v)
		}
		return v, err
	})
	wait := a.platformCommand("wait", "等待执行结束；只有 passed 返回退出码 0", projectScope, func(cmd *cobra.Command, _ []string, c *platform.Client, _ string) (any, error) {
		return execution.Wait(cmd.Context(), c, a.options.id, a.options.waitTimeout, a.options.poll)
	})
	a.options.waitFlags(wait)
	wait.PreRunE = a.validateWait
	cancel := a.platformCommand("cancel", "请求取消已有执行", projectScope, func(cmd *cobra.Command, _ []string, c *platform.Client, _ string) (any, error) {
		return c.Call(cmd.Context(), "execution", "CancelRun", jsonutil.Merge(c.Scope(), map[string]any{"id": a.options.id}))
	})
	for _, cmd := range []*cobra.Command{get, wait, cancel} {
		a.options.idFlag(cmd)
		_ = cmd.MarkFlagRequired("id")
		cmd.Args = a.requireRunID
	}
	group.AddCommand(list, get, wait, cancel)
	return group
}

func (a *application) taskRunCommand() *cobra.Command {
	cmd := a.platformCommand("run", "启动测试任务，可等待执行结束", projectScope, func(cmd *cobra.Command, _ []string, c *platform.Client, _ string) (any, error) {
		return runTask(cmd.Context(), c, a.options, cmd.ErrOrStderr())
	})
	o := &a.options
	o.idFlag(cmd)
	o.waitFlags(cmd)
	cmd.Flags().BoolVar(&o.wait, "wait", false, "提交后等待测试结束")
	cmd.Flags().StringVar(&o.idempotency, "idempotency-key", "", "本次提交的幂等键；提交结果不明时使用原键重试")
	_ = cmd.MarkFlagRequired("id")
	cmd.Args = a.requireRunID
	cmd.PreRunE = a.validateWait
	cmd.Long = "使用任务已保存的环境和 Runner 执行。等待超时不会取消平台执行，可使用返回的执行 ID 继续查询。"
	cmd.Example = "  apimoon task run --id <任务ID> --idempotency-key <本次执行唯一键> --wait\n  apimoon run get --id <执行ID>"
	return cmd
}

func (a *application) requireRunID(cmd *cobra.Command, args []string) error {
	if err := cobra.NoArgs(cmd, args); err != nil {
		return err
	}
	if a.options.id == "" {
		return clierror.New("usage", "需要非空 --id")
	}
	return nil
}

func (a *application) validateWait(*cobra.Command, []string) error {
	if a.options.waitTimeout <= 0 || a.options.poll <= 0 {
		return clierror.New("usage", "等待超时和轮询间隔必须大于 0")
	}
	return nil
}

func runTask(ctx context.Context, c *platform.Client, o options, stderr io.Writer) (any, error) {
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
