package agent

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	m "github.com/mi991203/TestMoon/libs/go/testing/domain"
	"io"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/netip"
	"net/url"
	"strings"
	"time"
)

func publicIP(ip net.IP) bool {
	addr, ok := netip.AddrFromSlice(ip)
	if !ok {
		return false
	}
	addr = addr.Unmap()
	if !addr.IsGlobalUnicast() || addr.IsPrivate() || addr.IsLoopback() || addr.IsLinkLocalUnicast() {
		return false
	}
	for _, cidr := range []string{"0.0.0.0/8", "100.64.0.0/10", "192.0.0.0/24", "192.0.2.0/24", "198.18.0.0/15", "198.51.100.0/24", "203.0.113.0/24", "240.0.0.0/4", "2001:db8::/32", "2001::/32", "2002::/16", "::/96"} {
		if netip.MustParsePrefix(cidr).Contains(addr) {
			return false
		}
	}
	return true
}

type network struct {
	Public  bool
	Clients map[string]*http.Client
}

func (n *network) client(item string) *http.Client {
	if c := n.Clients[item]; c != nil {
		return c
	}
	jar, _ := cookiejar.New(nil)
	transport := &http.Transport{Proxy: nil, MaxResponseHeaderBytes: 64 << 10, DisableKeepAlives: true, TLSHandshakeTimeout: 10 * time.Second}
	transport.DialContext = func(ctx context.Context, network, address string) (net.Conn, error) {
		host, port, err := net.SplitHostPort(address)
		if err != nil {
			return nil, err
		}
		ips, err := net.DefaultResolver.LookupIPAddr(ctx, host)
		if err != nil {
			return nil, errors.New("目标DNS解析失败")
		}
		if len(ips) == 0 {
			return nil, errors.New("目标没有可用地址")
		}
		if n.Public {
			for _, ip := range ips {
				if !publicIP(ip.IP) {
					return nil, errors.New("公共Runner禁止访问私网、环回或保留地址")
				}
			}
		}
		var last error
		for _, ip := range ips {
			dialer := net.Dialer{Timeout: 10 * time.Second}
			conn, err := dialer.DialContext(ctx, network, net.JoinHostPort(ip.IP.String(), port))
			if err == nil {
				return conn, nil
			}
			last = err
		}
		return nil, last
	}
	client := &http.Client{Transport: transport, Jar: jar, CheckRedirect: func(r *http.Request, via []*http.Request) error {
		if len(via) >= 5 {
			return errors.New("重定向次数超限")
		}
		if r.URL.Scheme != "http" && r.URL.Scheme != "https" {
			return errors.New("不支持的重定向协议")
		}
		if r.URL.User != nil {
			return errors.New("URL中不允许认证信息")
		}
		return nil
	}}
	n.Clients[item] = client
	return client
}
func (n *network) call(ctx context.Context, item string, request m.HTTPRequest) m.Response {
	start := time.Now()
	out := m.Response{Headers: map[string][]string{}}
	fail := func(s string) m.Response {
		out.Error = s
		out.DurationMS = time.Since(start).Milliseconds()
		return out
	}
	if !strings.Contains("|GET|POST|PUT|PATCH|DELETE|HEAD|OPTIONS|", "|"+request.Method+"|") {
		return fail("不支持的HTTP方法")
	}
	u, e := url.Parse(request.URL)
	if e != nil || u.Host == "" || (u.Scheme != "http" && u.Scheme != "https") || u.User != nil || len(request.URL) > 8192 {
		return fail("仅允许无内嵌认证信息的HTTP/HTTPS URL")
	}
	if len(request.Body) > 1<<20 {
		return fail("请求体超过1MiB")
	}
	query := u.Query()
	for k, v := range request.Query {
		query.Set(k, v)
	}
	if request.Auth.Type == "api_key" && request.Auth.In == "query" {
		query.Set(request.Auth.Key, request.Auth.Value)
	}
	u.RawQuery = query.Encode()
	timeout := 30 * time.Second
	if request.TimeoutMS > 0 {
		timeout = time.Duration(min(request.TimeoutMS, 120000)) * time.Millisecond
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	r, e := http.NewRequestWithContext(ctx, request.Method, u.String(), strings.NewReader(request.Body))
	if e != nil {
		return fail("请求构造失败")
	}
	for k, v := range request.Headers {
		if strings.ContainsAny(k+v, "\r\n") {
			return fail("请求头包含非法换行")
		}
		if strings.EqualFold(k, "Host") {
			r.Host = v
		} else {
			r.Header.Set(k, v)
		}
	}
	switch request.Auth.Type {
	case "basic":
		r.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(request.Auth.Username+":"+request.Auth.Password)))
	case "bearer":
		r.Header.Set("Authorization", "Bearer "+request.Auth.Token)
	case "api_key":
		if request.Auth.In != "query" {
			r.Header.Set(request.Auth.Key, request.Auth.Value)
		}
	}
	if r.Header.Get("Content-Type") == "" {
		switch request.BodyType {
		case "json":
			r.Header.Set("Content-Type", "application/json")
		case "form":
			r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		case "text":
			r.Header.Set("Content-Type", "text/plain")
		}
	}
	response, e := n.client(item).Do(r)
	if e != nil {
		if ctx.Err() != nil {
			return fail("HTTP请求超时或已取消")
		}
		return fail(fmt.Sprintf("HTTP调用失败: %s", e.Error()))
	}
	defer response.Body.Close()
	out.Status = response.StatusCode
	out.Headers = response.Header.Clone()
	body, e := io.ReadAll(io.LimitReader(response.Body, (1<<20)+1))
	if e != nil {
		return fail("读取响应失败")
	}
	if len(body) > 1<<20 {
		out.Body = string(body[:1<<20])
		out.Truncated = true
		return fail("响应超过1MiB，已截断")
	}
	out.Body = string(body)
	out.DurationMS = time.Since(start).Milliseconds()
	return out
}
