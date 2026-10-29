package cli

import (
	"context"
	"io"

	"github.com/mi991203/TestMoon/tools/apimoon/clierror"
	"github.com/mi991203/TestMoon/tools/apimoon/jsonutil"
	"github.com/mi991203/TestMoon/tools/apimoon/platform"
	"github.com/mi991203/TestMoon/tools/apimoon/validate"
	"github.com/spf13/cobra"
)

func (a *application) resourceCommand(kind string) *cobra.Command {
	label := map[string]string{"api": "接口", "case": "接口用例", "scenario": "场景用例", "task": "测试任务", "environment": "环境", "folder": "目录"}[kind]
	group := commandGroup(kind, "管理"+label)
	o := &a.options
	for _, action := range []string{"list", "get", "create", "update", "delete"} {
		verb := map[string]string{"list": "列出", "get": "查看", "create": "创建", "update": "更新", "delete": "删除"}[action]
		cmd := a.platformCommand(action, verb+label, projectScope, func(cmd *cobra.Command, _ []string, c *platform.Client, _ string) (any, error) {
			return resourceCommand(cmd.Context(), c, *o, kind, action, cmd.InOrStdin())
		})
		o.inputFlag(cmd)
		switch action {
		case "list":
			if kind != "task" {
				cmd.Flags().StringVar(&o.search, "search", "", "按名称搜索")
				cmd.Flags().StringVar(&o.folder, "folder-id", "", "按目录 ID 筛选，显式空值选择未分组资源")
			}
			if kind == "case" {
				cmd.Flags().StringVar(&o.api, "api-id", "", "按接口 ID 筛选")
			}
		case "get", "delete":
			o.idFlag(cmd)
			cmd.MarkFlagsOneRequired("id", "file")
		case "create":
			_ = cmd.MarkFlagRequired("file")
			if kind != "task" {
				cmd.Flags().StringVar(&o.folder, "folder-id", "", "所属目录 ID")
			}
			if kind == "api" || kind == "case" || kind == "task" {
				o.defaultFlags(cmd)
			}
		case "update":
			_ = cmd.MarkFlagRequired("file")
			o.idFlag(cmd)
			cmd.Flags().StringVar(&o.revision, "revision", "", "更新所依据的资源版本")
			if kind != "task" {
				cmd.Flags().StringVar(&o.folder, "folder-id", "", "所属目录 ID")
			}
			cmd.Long = "更新资源，需要 ID、revision、名称和完整 document；可在输入 JSON 中提供。不会局部合并文档。"
		}
		group.AddCommand(cmd)
	}
	if kind == "task" {
		group.AddCommand(a.taskRunCommand())
	}
	return group
}

func resourceCommand(ctx context.Context, c *platform.Client, o options, group, action string, stdin io.Reader) (any, error) {
	input := map[string]any{}
	if o.file != "" {
		if err := readInput(o.file, stdin, &input); err != nil {
			return nil, err
		}
		if input == nil {
			return nil, clierror.New("validation", "输入必须是 JSON 对象")
		}
	}
	for key, value := range map[string]string{"id": o.id, "revision": o.revision, "search": o.search, "folder_id": o.folder, "api_id": o.api} {
		flagKey := map[string]string{"folder_id": "folder-id", "api_id": "api-id"}[key]
		if flagKey == "" {
			flagKey = key
		}
		if o.set[flagKey] {
			input[key] = value
		}
	}
	if action == "create" {
		c.Defaults(group, input)
	} else if action == "update" && (o.set["environment"] || o.set["runner"]) {
		return nil, clierror.New("usage", "更新请在完整 document 中明确填写环境和 Runner")
	}
	scoped, err := c.Scoped(input)
	if err != nil {
		return nil, err
	}
	if err = validate.Asset(group, action, scoped); err != nil {
		return nil, err
	}
	v, err := c.Resource(ctx, group, action, input)
	if err == nil {
		if r := jsonutil.Object(v["resource"]); r != nil {
			v["url"] = c.Link(group, jsonutil.String(r["id"]))
		}
	}
	return v, err
}
