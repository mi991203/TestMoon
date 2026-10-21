package api

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/mi991203/TestMoon/backend/testing/mock/model"
	p "github.com/mi991203/TestMoon/libs/go/platform"
)

type ruleInput struct {
	EndpointID string         `json:"endpoint_id" validate:"required"`
	Name       string         `json:"name" validate:"required,max=128"`
	Method     string         `json:"method" validate:"oneof=GET POST PUT PATCH DELETE HEAD OPTIONS"`
	Path       string         `json:"path" validate:"required,max=1024"`
	Priority   int            `json:"priority"`
	Document   model.Response `json:"document"`
}

func validateRuleDefinition(path string, response model.Response) error {
	if !strings.HasPrefix(path, "/") || response.Status < 200 || response.Status > 599 || response.DelayMS < 0 || response.DelayMS > 30000 || len(response.Body) > 1<<20 || len(response.Conditions) > 20 {
		return p.Bad("Mock路径、状态码、延迟或内容超出限制")
	}
	for _, c := range response.Conditions {
		if c.Source != "query" && c.Source != "header" && c.Source != "body" {
			return p.Bad("不支持的匹配来源")
		}
		if c.Operator != "eq" && c.Operator != "contains" && c.Operator != "exists" && c.Operator != "regex" {
			return p.Bad("不支持的条件运算符")
		}
		if c.Operator == "regex" {
			if len(fmt.Sprint(c.Expected)) > 512 {
				return p.Bad("正则过长")
			}
			if _, e := regexp.Compile(fmt.Sprint(c.Expected)); e != nil {
				return p.Bad("无效正则")
			}
		}
	}
	for k, v := range response.Headers {
		if strings.ContainsAny(k+v, "\r\n") || !regexp.MustCompile("^[A-Za-z0-9!#$%&'*+.^_`|~-]+$").MatchString(k) {
			return p.Bad("无效响应头")
		}
		switch strings.ToLower(k) {
		case "connection", "transfer-encoding", "content-length", "upgrade", "trailer", "keep-alive", "proxy-authenticate", "proxy-authorization":
			return p.Bad("不允许配置逐跳响应头")
		}
	}
	return nil
}
