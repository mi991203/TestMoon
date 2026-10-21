package validation

import (
	"strings"

	p "github.com/mi991203/TestMoon/libs/go/platform"
	m "github.com/mi991203/TestMoon/libs/go/testing/domain"
)

func Request(r m.HTTPRequest) error {
	if r.URL == "" || len(r.URL) > 8192 {
		return p.Bad("请求URL不能为空或过长")
	}
	if !strings.Contains("|GET|POST|PUT|PATCH|DELETE|HEAD|OPTIONS|", "|"+strings.ToUpper(r.Method)+"|") {
		return p.Bad("不支持的HTTP方法")
	}
	if len(r.Body) > 1<<20 || r.TimeoutMS < 0 || r.TimeoutMS > 120000 {
		return p.Bad("请求体或超时超出限制")
	}
	if r.BodyType != "" && r.BodyType != "json" && r.BodyType != "text" && r.BodyType != "form" {
		return p.Bad("仅支持JSON、文本和URL编码表单")
	}
	if !strings.Contains("||none|basic|bearer|api_key|", "|"+r.Auth.Type+"|") {
		return p.Bad("不支持的认证类型")
	}
	return nil
}
