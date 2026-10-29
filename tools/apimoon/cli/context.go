package cli

import (
	"github.com/mi991203/TestMoon/tools/apimoon/clierror"
	"github.com/mi991203/TestMoon/tools/apimoon/config"
	"github.com/mi991203/TestMoon/tools/apimoon/platform"
	"github.com/spf13/cobra"
)

type requiredScope int

const (
	noScope requiredScope = iota
	orgScope
	projectScope
)

// Load configuration only for commands that need it.
func (a *application) configPath() (string, error) {
	if a.options.config != "" {
		return a.options.config, nil
	}
	return config.DefaultPath()
}

func (a *application) platformCommand(use, short string, scope requiredScope, run func(*cobra.Command, []string, *platform.Client, string) (any, error)) *cobra.Command {
	return a.command(use, short, func(cmd *cobra.Command, args []string) (any, error) {
		c, file, err := a.client(scope)
		if err != nil {
			return nil, err
		}
		return run(cmd, args, c, file)
	})
}

func (a *application) client(scope requiredScope) (*platform.Client, string, error) {
	file, err := a.configPath()
	if err != nil {
		return nil, "", err
	}
	c, err := config.Load(file)
	if err != nil {
		return nil, file, clierror.New("config", "无法读取配置，请先执行 config init")
	}
	if err = c.Validate(); err != nil {
		return nil, file, err
	}
	o := a.options
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
	if scope >= orgScope && c.Context.OrgID == "" {
		return nil, file, clierror.New("config", "缺少 org_id，请配置 context.org_id 或 --org")
	}
	if scope >= projectScope && c.Context.ProjectID == "" {
		return nil, file, clierror.New("config", "缺少 project_id，请配置 context.project_id 或 --project")
	}
	return platform.New(c), file, nil
}

func (a *application) contextCommand() *cobra.Command {
	cmd := commandGroup("context", "查看当前作用域")
	cmd.AddCommand(a.platformCommand("show", "查看当前组织、项目、环境和 Runner", noScope, func(_ *cobra.Command, _ []string, c *platform.Client, _ string) (any, error) {
		return c.Config.Context, nil
	}))
	return cmd
}

func (a *application) orgCommand() *cobra.Command {
	cmd := commandGroup("org", "查询组织")
	cmd.AddCommand(a.platformCommand("list", "列出当前用户的组织", noScope, func(cmd *cobra.Command, _ []string, c *platform.Client, _ string) (any, error) {
		return c.Call(cmd.Context(), "identity", "ListOrganizations", map[string]any{})
	}))
	return cmd
}

func (a *application) projectCommand() *cobra.Command {
	cmd := commandGroup("project", "查询项目")
	cmd.AddCommand(a.platformCommand("list", "列出当前组织的项目", orgScope, func(cmd *cobra.Command, _ []string, c *platform.Client, _ string) (any, error) {
		return c.Call(cmd.Context(), "core", "ListProjects", map[string]any{"org_id": c.Config.Context.OrgID})
	}))
	return cmd
}

func (a *application) runnerCommand() *cobra.Command {
	cmd := commandGroup("runner", "查询平台 Runner")
	cmd.Long = "查询当前项目在 TestMoon 平台上的 Runner；本机 Runner 使用独立的 testmoon-runner 程序启动。"
	cmd.AddCommand(a.platformCommand("list", "列出当前项目的 Runner 及在线状态", projectScope, func(cmd *cobra.Command, _ []string, c *platform.Client, _ string) (any, error) {
		return c.Call(cmd.Context(), "execution", "ListProjectRunners", c.Scope())
	}))
	return cmd
}
