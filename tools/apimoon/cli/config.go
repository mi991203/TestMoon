package cli

import (
	"errors"
	"os"
	"strconv"

	"github.com/mi991203/TestMoon/tools/apimoon/clierror"
	"github.com/mi991203/TestMoon/tools/apimoon/config"
	"github.com/spf13/cobra"
)

func (a *application) configCommand() *cobra.Command {
	group := commandGroup("config", "初始化和管理本机配置")
	init := a.command("init", "初始化配置文件，不覆盖已有配置", func(*cobra.Command, []string) (any, error) {
		file, err := a.configPath()
		if err != nil {
			return nil, err
		}
		return initConfig(file, a.options.base)
	})
	init.Flags().StringVar(&a.options.base, "base-url", "", "TestMoon 网关根地址，例如 http://localhost:18080")
	_ = init.MarkFlagRequired("base-url")
	show := a.command("show", "查看配置，隐藏登录凭证", func(*cobra.Command, []string) (any, error) {
		c, _, err := a.readConfig()
		if err != nil {
			return nil, err
		}
		c.Auth.CookieValue = "***"
		c.Auth.CSRF = "***"
		return c, nil
	})
	set := a.command("set KEY VALUE", "修改指定配置项", func(_ *cobra.Command, args []string) (any, error) {
		c, file, err := a.readConfig()
		if err != nil {
			return nil, err
		}
		return setConfig(file, c, args[0], args[1])
	})
	set.Args = cobra.ExactArgs(2)
	set.Example = "  apimoon config set context.project_id <项目ID>\n  apimoon config set context.runner_id -- \"\""
	group.AddCommand(init, show, set)
	return group
}

func (a *application) readConfig() (config.Config, string, error) {
	file, err := a.configPath()
	if err != nil {
		return config.Config{}, "", err
	}
	c, err := config.Load(file)
	if err != nil {
		return c, file, clierror.New("config", "无法读取配置，请先执行 config init")
	}
	return c, file, nil
}

func initConfig(file, rawURL string) (any, error) {
	if _, err := os.Stat(file); !errors.Is(err, os.ErrNotExist) {
		return nil, clierror.New("conflict", "配置已存在或无法访问，拒绝覆盖")
	}
	base, err := config.NormalizeURL(rawURL)
	if err != nil {
		return nil, err
	}
	c := config.Config{BaseURL: base, HTTPTimeoutSeconds: 30}
	if err = config.Save(file, c); err != nil {
		return nil, err
	}
	return map[string]any{"config_file": file}, nil
}

func setConfig(file string, c config.Config, key, value string) (any, error) {
	switch key {
	case "base_url":
		normalized, err := config.NormalizeURL(value)
		if err != nil {
			return nil, err
		}
		if normalized != c.BaseURL {
			c.Auth = config.AuthConfig{}
		}
		c.BaseURL = normalized
	case "context.org_id":
		if c.Context.OrgID != value {
			c.Context.ProjectID = ""
			c.Context.EnvironmentID = ""
			c.Context.RunnerID = ""
		}
		c.Context.OrgID = value
	case "context.project_id":
		if c.Context.ProjectID != value {
			c.Context.EnvironmentID = ""
			c.Context.RunnerID = ""
		}
		c.Context.ProjectID = value
	case "context.environment_id":
		c.Context.EnvironmentID = value
	case "context.runner_id":
		c.Context.RunnerID = value
	case "http_timeout_seconds":
		n, err := strconv.Atoi(value)
		if err != nil {
			return nil, clierror.New("config", "超时必须是整数秒")
		}
		c.HTTPTimeoutSeconds = n
	default:
		return nil, clierror.New("config", "未知或不可修改的配置项")
	}
	if err := config.Save(file, c); err != nil {
		return nil, err
	}
	return map[string]any{"updated": key}, nil
}
