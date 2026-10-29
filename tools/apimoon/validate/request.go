package validate

import (
	"slices"

	domain "github.com/mi991203/TestMoon/libs/go/testing/domain"
	"github.com/mi991203/TestMoon/tools/apimoon/clierror"
)

func validateRequest(r domain.HTTPRequest) error {
	if r.URL == "" || len(r.URL) > 8192 {
		return clierror.New("validation", "请求 URL 不能为空或超过 8192 字节")
	}
	if !slices.Contains([]string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"}, r.Method) {
		return clierror.New("validation", "不支持的 HTTP 方法")
	}
	if len(r.Body) > 1<<20 || r.TimeoutMS < 0 || r.TimeoutMS > 120000 {
		return clierror.New("validation", "请求体或超时超出平台限制")
	}
	if !slices.Contains([]string{"", "json", "text", "form"}, r.BodyType) || !slices.Contains([]string{"", "none", "basic", "bearer", "api_key"}, r.Auth.Type) {
		return clierror.New("validation", "不支持的请求体或认证类型")
	}
	return nil
}
func validatePatch(p domain.RequestPatch) error {
	if p.Method != nil && !slices.Contains([]string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"}, *p.Method) {
		return clierror.New("validation", "覆盖包含无效 HTTP 方法")
	}
	if p.URL != nil && (*p.URL == "" || len(*p.URL) > 8192) {
		return clierror.New("validation", "覆盖包含无效 URL")
	}
	if p.BodyType != nil && !slices.Contains([]string{"", "json", "text", "form"}, *p.BodyType) {
		return clierror.New("validation", "覆盖包含无效 body_type")
	}
	if p.Body != nil && len(*p.Body) > 1<<20 {
		return clierror.New("validation", "覆盖请求体超过 1MiB")
	}
	if p.Auth != nil && !slices.Contains([]string{"", "none", "basic", "bearer", "api_key"}, p.Auth.Type) {
		return clierror.New("validation", "覆盖包含无效认证类型")
	}
	return nil
}
