package platform

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/mi991203/TestMoon/tools/apimoon/clierror"
	"github.com/mi991203/TestMoon/tools/apimoon/config"
	"github.com/mi991203/TestMoon/tools/apimoon/jsonutil"
	"google.golang.org/genproto/googleapis/api/annotations"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/dynamicpb"
)

type Client struct {
	Config config.Config
	HTTP   *http.Client
}

func New(c config.Config) *Client {
	return &Client{Config: c, HTTP: &http.Client{Timeout: time.Duration(c.HTTPTimeoutSeconds) * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}}
}

func (c *Client) Call(ctx context.Context, service, name string, input map[string]any) (map[string]any, error) {
	v, _, err := c.CallWithCookies(ctx, service, name, input)
	return v, err
}
func (c *Client) CallWithCookies(ctx context.Context, service, name string, input map[string]any) (map[string]any, []*http.Cookie, error) {
	base, err := config.NormalizeURL(c.Config.BaseURL)
	if err != nil {
		return nil, nil, err
	}
	b, err := CanonicalRequest(service, name, input)
	if err != nil {
		return nil, nil, err
	}
	anonymous := service == "identity" && name == "Login"
	if !anonymous {
		a := c.Config.Auth
		if a.BaseURL != base || a.CookieValue == "" || a.CSRF == "" || (a.CookieName != "testmoon_session" && a.CookieName != "__Host-testmoon_session") {
			return nil, nil, clierror.New("auth", "缺少当前服务的登录凭证，请运行 auth login")
		}
		if !a.ExpiresAt.IsZero() && time.Now().After(a.ExpiresAt) {
			return nil, nil, clierror.New("auth", "登录已过期，请重新运行 auth login")
		}
	}
	m := Method(service, name)
	rule := proto.GetExtension(m.Options(), annotations.E_Http).(*annotations.HttpRule)
	req, err := http.NewRequestWithContext(ctx, "POST", base+rule.GetPost(), bytes.NewReader(b))
	if err != nil {
		return nil, nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if !anonymous {
		req.AddCookie(&http.Cookie{Name: c.Config.Auth.CookieName, Value: c.Config.Auth.CookieValue})
		req.Header.Set("X-CSRF-Token", c.Config.Auth.CSRF)
	}
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, nil, clierror.New("transport", "平台请求未取得响应；写操作的结果可能不明")
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(io.LimitReader(resp.Body, 16<<20+1))
	if err != nil {
		return nil, nil, clierror.New("transport", "平台响应读取失败；写操作的结果可能不明")
	}
	if len(raw) > 16<<20 {
		return nil, nil, clierror.New("protocol", "平台响应超过 16MiB")
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		code := "platform"
		switch resp.StatusCode {
		case 401:
			code = "auth"
		case 403:
			code = "permission"
		case 409:
			code = "conflict"
		case 400:
			code = "validation"
		}
		message := strings.TrimSpace(string(raw))
		var v map[string]any
		if json.Unmarshal(raw, &v) == nil && v["message"] != nil {
			message = jsonutil.String(v["message"])
		}
		if len(message) > 2048 {
			message = message[:2048]
		}
		message = clierror.Redact(message, c.Config.Auth.CookieValue, c.Config.Auth.CSRF, jsonutil.String(input["password"]))
		return nil, nil, &clierror.Error{Code: code, Message: message, Status: resp.StatusCode}
	}
	p := dynamicpb.NewMessage(m.Output())
	if err = protojson.Unmarshal(raw, p); err != nil {
		return nil, nil, clierror.New("protocol", "平台响应不符合当前协议；写操作结果需核对")
	}
	raw, err = (protojson.MarshalOptions{UseProtoNames: true, EmitUnpopulated: true}).Marshal(p)
	if err != nil {
		return nil, nil, err
	}
	var result map[string]any
	if err = jsonutil.Decode(raw, &result); err != nil {
		return nil, nil, err
	}
	return result, resp.Cookies(), nil
}
