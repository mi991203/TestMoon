package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/mi991203/TestMoon/tools/apimoon/clierror"
	"github.com/mi991203/TestMoon/tools/apimoon/config"
	"github.com/mi991203/TestMoon/tools/apimoon/plan"
	"github.com/mi991203/TestMoon/tools/apimoon/platform"
)

func Execute(ctx context.Context, args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	o, pos, err := parse(args)
	if err == nil && (o.help || len(pos) == 0) {
		if o.output == "text" {
			fmt.Fprint(stdout, help)
		} else {
			_ = json.NewEncoder(stdout).Encode(map[string]any{"ok": true, "data": map[string]any{"help": help}})
		}
		return 0
	}
	var data any
	if err == nil {
		data, err = dispatch(ctx, o, pos, stdin, stderr)
	}
	out := map[string]any{"ok": err == nil}
	if data != nil {
		out["data"] = data
	}
	code := 0
	if err != nil {
		e := clierror.From(err)
		out["error"] = e
		switch e.Code {
		case "usage", "validation", "config":
			code = 2
		case "auth", "permission":
			code = 3
		case "conflict", "uncertain":
			code = 4
		case "wait_timeout":
			code = 5
		case "execution":
			code = 6
		default:
			code = 1
		}
	}
	if o.output == "text" {
		b, _ := json.MarshalIndent(out, "", "  ")
		fmt.Fprintln(stdout, string(b))
	} else {
		if e := json.NewEncoder(stdout).Encode(out); e != nil {
			return 1
		}
	}
	return code
}
func dispatch(ctx context.Context, o options, pos []string, stdin io.Reader, stderr io.Writer) (any, error) {
	if len(pos) < 2 {
		return nil, clierror.New("usage", "需要命令和子命令，使用 --help 查看")
	}
	group, action := pos[0], pos[1]
	file := o.config
	if file == "" {
		var err error
		file, err = config.DefaultPath()
		if err != nil {
			return nil, err
		}
	}
	if group == "plan" && action == "schema" {
		return plan.Schema(), nil
	}
	if group == "config" {
		return configCommand(file, o, pos)
	}
	c, err := config.Load(file)
	if err != nil {
		return nil, clierror.New("config", "无法读取配置，请先执行 config init")
	}
	if err = c.Validate(); err != nil {
		return nil, err
	}
	if o.set["org"] {
		if c.Context.OrgID != o.org {
			c.Context.ProjectID = ""
			c.Context.EnvironmentID = ""
			c.Context.RunnerID = ""
		}
		c.Context.OrgID = o.org
	}
	if o.set["project"] {
		if c.Context.ProjectID != o.project {
			c.Context.EnvironmentID = ""
			c.Context.RunnerID = ""
		}
		c.Context.ProjectID = o.project
	}
	if o.set["environment"] {
		c.Context.EnvironmentID = o.environment
	}
	if o.set["runner"] {
		c.Context.RunnerID = o.runner
	}
	client := platform.New(c)
	if group == "context" && action == "show" {
		return c.Context, nil
	}
	if group == "auth" {
		return authenticate(ctx, client, file, o, action, stdin)
	}
	if group == "org" && action == "list" {
		return client.Call(ctx, "identity", "ListOrganizations", map[string]any{})
	}
	if c.Context.OrgID == "" {
		return nil, clierror.New("config", "缺少 org_id，请配置 context.org_id 或 --org")
	}
	if group == "project" && action == "list" {
		return client.Call(ctx, "core", "ListProjects", map[string]any{"org_id": c.Context.OrgID})
	}
	if c.Context.ProjectID == "" {
		return nil, clierror.New("config", "缺少 project_id，请配置 context.project_id 或 --project")
	}
	if group == "runner" && action == "list" {
		return client.Call(ctx, "execution", "ListProjectRunners", client.Scope())
	}
	if group == "plan" {
		return planCommand(ctx, client, file, o, action, stdin)
	}
	if group == "run" || (group == "task" && action == "run") {
		return runCommand(ctx, client, o, group, action, stderr)
	}
	if _, ok := platform.Types[group]; !ok {
		return nil, clierror.New("usage", "未知命令组")
	}
	return resourceCommand(ctx, client, o, group, action, stdin)
}
