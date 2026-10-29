package cli

import (
	"errors"
	"os"
	"strconv"

	"github.com/mi991203/TestMoon/tools/apimoon/clierror"
	"github.com/mi991203/TestMoon/tools/apimoon/config"
)

func configCommand(file string, o options, pos []string) (any, error) {
	action := pos[1]
	if action == "init" {
		if _, err := os.Stat(file); !errors.Is(err, os.ErrNotExist) {
			return nil, clierror.New("conflict", "配置已存在或无法访问，拒绝覆盖")
		}
		base, err := config.NormalizeURL(o.base)
		if err != nil {
			return nil, err
		}
		c := config.Config{BaseURL: base, HTTPTimeoutSeconds: 30}
		if err = config.Save(file, c); err != nil {
			return nil, err
		}
		return map[string]any{"config_file": file}, nil
	}
	c, err := config.Load(file)
	if err != nil {
		return nil, clierror.New("config", "无法读取配置，请先执行 config init")
	}
	switch action {
	case "show":
		c.Auth.CookieValue = "***"
		c.Auth.CSRF = "***"
		return c, nil
	case "set":
		if len(pos) != 4 {
			return nil, clierror.New("usage", "config set KEY VALUE")
		}
		k, v := pos[2], pos[3]
		switch k {
		case "base_url":
			normalized, e := config.NormalizeURL(v)
			if e != nil {
				return nil, e
			}
			if normalized != c.BaseURL {
				c.Auth = config.AuthConfig{}
			}
			c.BaseURL = normalized
		case "context.org_id":
			if c.Context.OrgID != v {
				c.Context.ProjectID = ""
				c.Context.EnvironmentID = ""
				c.Context.RunnerID = ""
			}
			c.Context.OrgID = v
		case "context.project_id":
			if c.Context.ProjectID != v {
				c.Context.EnvironmentID = ""
				c.Context.RunnerID = ""
			}
			c.Context.ProjectID = v
		case "context.environment_id":
			c.Context.EnvironmentID = v
		case "context.runner_id":
			c.Context.RunnerID = v
		case "http_timeout_seconds":
			n, e := strconv.Atoi(v)
			if e != nil {
				return nil, clierror.New("config", "超时必须是整数秒")
			}
			c.HTTPTimeoutSeconds = n
		default:
			return nil, clierror.New("config", "未知或不可修改的配置项")
		}
		if err = config.Save(file, c); err != nil {
			return nil, err
		}
		return map[string]any{"updated": k}, nil
	default:
		return nil, clierror.New("usage", "未知 config 子命令")
	}
}
