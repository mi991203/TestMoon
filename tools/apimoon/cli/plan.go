package cli

import (
	"context"

	"github.com/mi991203/TestMoon/tools/apimoon/clierror"
	"github.com/mi991203/TestMoon/tools/apimoon/plan"
	"github.com/mi991203/TestMoon/tools/apimoon/platform"
	"github.com/spf13/cobra"
)

func (a *application) planCommand() *cobra.Command {
	group := commandGroup("plan", "校验、同步和恢复测试计划")
	schema := a.command("schema", "离线输出完整计划 JSON Schema", func(*cobra.Command, []string) (any, error) {
		return plan.Schema(), nil
	})
	validate := a.planFileCommand("validate", "校验计划及平台引用，不执行请求或写入资源", func(cmd *cobra.Command, c *platform.Client, _ plan.Syncer, p plan.Plan) (any, error) {
		return validatePlan(cmd.Context(), c, p)
	})
	a.options.defaultFlags(validate)
	sync := a.planFileCommand("sync", "按引用顺序同步计划，保留进度以便重试", func(cmd *cobra.Command, _ *platform.Client, s plan.Syncer, p plan.Plan) (any, error) {
		return syncPlan(cmd.Context(), s, p)
	})
	sync.Example = "  apimoon plan sync --file test-plan.json\n  apimoon plan sync --file - < test-plan.json"
	a.options.defaultFlags(sync)
	status := a.planFileCommand("status", "读取计划的本地同步状态", func(_ *cobra.Command, _ *platform.Client, s plan.Syncer, p plan.Plan) (any, error) {
		return s.Status(p.PlanID)
	})
	bind := a.planFileCommand("bind", "核对资源内容后恢复结果不明的写入绑定", func(cmd *cobra.Command, _ *platform.Client, s plan.Syncer, p plan.Plan) (any, error) {
		return s.Bind(cmd.Context(), p, a.options.key, a.options.id)
	})
	bind.Args = func(cmd *cobra.Command, args []string) error {
		if err := cobra.NoArgs(cmd, args); err != nil {
			return err
		}
		if a.options.key == "" || a.options.id == "" {
			return clierror.New("usage", "plan bind 需要非空 --key 和 --id")
		}
		return nil
	}
	bind.Flags().StringVar(&a.options.key, "key", "", "计划内条目的稳定 key")
	a.options.idFlag(bind)
	a.options.defaultFlags(bind)
	_ = bind.MarkFlagRequired("key")
	_ = bind.MarkFlagRequired("id")
	group.AddCommand(schema, validate, sync, status, bind)
	return group
}

func (a *application) planFileCommand(use, short string, run func(*cobra.Command, *platform.Client, plan.Syncer, plan.Plan) (any, error)) *cobra.Command {
	cmd := a.platformCommand(use, short, projectScope, func(cmd *cobra.Command, _ []string, c *platform.Client, file string) (any, error) {
		var p plan.Plan
		if err := readInput(a.options.file, cmd.InOrStdin(), &p); err != nil {
			return nil, err
		}
		if _, err := plan.Order(p); err != nil {
			return nil, err
		}
		return run(cmd, c, plan.Syncer{Client: c, ConfigFile: file}, p)
	})
	a.options.inputFlag(cmd)
	_ = cmd.MarkFlagRequired("file")
	return cmd
}

func validatePlan(ctx context.Context, c *platform.Client, p plan.Plan) (any, error) {
	entries, err := plan.Validate(ctx, c, p)
	if err != nil {
		return nil, err
	}
	keys := []string{}
	for _, e := range entries {
		keys = append(keys, e.Key)
	}
	return map[string]any{"plan_id": p.PlanID, "order": keys, "warnings": []string{"静态预校验不执行脚本或请求；动态变量、断言正确性及部署配置需运行验证"}}, nil
}

func syncPlan(ctx context.Context, s plan.Syncer, p plan.Plan) (any, error) {
	state, err := s.Sync(ctx, p)
	links := map[string]string{}
	for k, b := range state.Entries {
		if b.State == "confirmed" {
			links[k] = s.Client.Link(b.Kind, b.ID)
		}
	}
	return map[string]any{"state": state, "urls": links}, err
}
