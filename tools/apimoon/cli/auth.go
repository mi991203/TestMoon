package cli

import (
	"context"
	"io"
	"slices"
	"strings"
	"time"

	"github.com/mi991203/TestMoon/tools/apimoon/clierror"
	"github.com/mi991203/TestMoon/tools/apimoon/config"
	"github.com/mi991203/TestMoon/tools/apimoon/jsonutil"
	"github.com/mi991203/TestMoon/tools/apimoon/platform"
)

func authenticate(ctx context.Context, c *platform.Client, file string, o options, action string, stdin io.Reader) (any, error) {
	switch action {
	case "login":
		if !o.passwordStdin || o.username == "" {
			return nil, clierror.New("usage", "auth login 需要 --username USER --password-stdin")
		}
		b, err := io.ReadAll(io.LimitReader(stdin, 130))
		if err != nil {
			return nil, err
		}
		password := strings.TrimSuffix(strings.TrimSuffix(string(b), "\n"), "\r")
		if password == "" || len(password) > 128 {
			return nil, clierror.New("validation", "密码应为 1..128 字节")
		}
		v, cookies, err := c.CallWithCookies(ctx, "identity", "Login", map[string]any{"username": o.username, "password": password})
		if err != nil {
			return nil, err
		}
		base, _ := config.NormalizeURL(c.Config.BaseURL)
		var a config.AuthConfig
		for _, cookie := range cookies {
			if slices.Contains([]string{"testmoon_session", "__Host-testmoon_session"}, cookie.Name) && cookie.Value != "" {
				expiry := cookie.Expires
				if cookie.MaxAge > 0 {
					expiry = time.Now().Add(time.Duration(cookie.MaxAge) * time.Second)
				}
				a = config.AuthConfig{BaseURL: base, CookieName: cookie.Name, CookieValue: cookie.Value, CSRF: jsonutil.String(v["csrf"]), UserID: jsonutil.String(jsonutil.Object(v["user"])["id"]), ExpiresAt: expiry}
			}
		}
		if a.CookieValue == "" || a.CSRF == "" || a.UserID == "" {
			return nil, clierror.New("protocol", "登录响应未包含完整会话凭证")
		}
		c.Config.Auth = a
		if err = config.Save(file, c.Config); err != nil {
			return nil, err
		}
		return map[string]any{"user": v["user"], "expires_at": a.ExpiresAt}, nil
	case "status":
		v, err := c.Call(ctx, "identity", "Session", map[string]any{})
		if err != nil {
			return nil, err
		}
		delete(v, "csrf")
		delete(jsonutil.Object(v["principal"]), "csrf")
		return v, nil
	case "logout":
		v, err := c.Call(ctx, "identity", "Logout", map[string]any{})
		if err != nil && clierror.From(err).Code != "auth" {
			return nil, err
		}
		c.Config.Auth = config.AuthConfig{}
		if err = config.Save(file, c.Config); err != nil {
			return nil, err
		}
		if v == nil {
			v = map[string]any{"ok": true}
		}
		return v, nil
	default:
		return nil, clierror.New("usage", "未知 auth 子命令")
	}
}
