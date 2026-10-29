package config

import (
	"bytes"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/mi991203/TestMoon/tools/apimoon/clierror"
	"github.com/mi991203/TestMoon/tools/apimoon/fileio"
	"go.yaml.in/yaml/v3"
)

type Config struct {
	BaseURL            string     `yaml:"base_url" json:"base_url"`
	Auth               AuthConfig `yaml:"auth" json:"auth"`
	Context            Scope      `yaml:"context" json:"context"`
	HTTPTimeoutSeconds int        `yaml:"http_timeout_seconds" json:"http_timeout_seconds"`
}
type AuthConfig struct {
	BaseURL     string    `yaml:"base_url,omitempty" json:"base_url,omitempty"`
	CookieName  string    `yaml:"cookie_name,omitempty" json:"cookie_name,omitempty"`
	CookieValue string    `yaml:"cookie_value,omitempty" json:"cookie_value,omitempty"`
	CSRF        string    `yaml:"csrf,omitempty" json:"csrf,omitempty"`
	UserID      string    `yaml:"user_id,omitempty" json:"user_id,omitempty"`
	ExpiresAt   time.Time `yaml:"expires_at,omitempty" json:"expires_at,omitempty"`
}
type Scope struct {
	OrgID         string `yaml:"org_id" json:"org_id"`
	ProjectID     string `yaml:"project_id" json:"project_id"`
	EnvironmentID string `yaml:"environment_id,omitempty" json:"environment_id,omitempty"`
	RunnerID      string `yaml:"runner_id,omitempty" json:"runner_id,omitempty"`
}

func DefaultPath() (string, error) {
	dir, err := os.UserHomeDir()
	return filepath.Join(dir, ".api_moon", "config.yaml"), err
}
func Load(file string) (Config, error) {
	c := Config{HTTPTimeoutSeconds: 30}
	b, err := os.ReadFile(file)
	if err != nil {
		return c, err
	}
	d := yaml.NewDecoder(bytes.NewReader(b))
	d.KnownFields(true)
	if err = d.Decode(&c); err != nil {
		return c, clierror.New("config", "配置格式不正确或包含未知字段")
	}
	var extra any
	if err = d.Decode(&extra); err != io.EOF {
		return c, clierror.New("config", "配置必须是单份 YAML 文档")
	}
	return c, nil
}
func NormalizeURL(raw string) (string, error) {
	u, err := url.Parse(raw)
	if err != nil || u.Host == "" || (u.Scheme != "http" && u.Scheme != "https") || u.User != nil || u.RawQuery != "" || u.Fragment != "" || (u.Path != "" && u.Path != "/") {
		return "", clierror.New("config", "base_url 必须是 http(s) 网关根地址，不含账号、路径或查询参数")
	}
	u.Path = ""
	return u.String(), nil
}
func (c Config) Validate() error {
	if _, err := NormalizeURL(c.BaseURL); err != nil {
		return err
	}
	if c.HTTPTimeoutSeconds < 1 || c.HTTPTimeoutSeconds > 300 {
		return clierror.New("config", "http_timeout_seconds 应为 1..300")
	}
	return nil
}
func Save(file string, c Config) error {
	if err := c.Validate(); err != nil {
		return err
	}
	b, err := yaml.Marshal(c)
	if err != nil {
		return err
	}
	return fileio.WriteAtomic(file, b)
}
