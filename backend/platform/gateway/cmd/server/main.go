package main

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"fmt"
	p "github.com/mi991203/TestMoon/libs/go/platform"
	iam "github.com/mi991203/TestMoon/protocols/platform/identity/v1"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

type gateway struct {
	C        p.Config
	Identity iam.IdentityClient
	Proxies  map[string]*httputil.ReverseProxy
	Mock     *httputil.ReverseProxy
	Root     string
}

func proxy(target string) *httputil.ReverseProxy {
	u, e := url.Parse(target)
	if e != nil {
		panic(e)
	}
	proxy := httputil.NewSingleHostReverseProxy(u)
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, e error) { http.Error(w, "upstream unavailable", 503) }
	return proxy
}
func main() {
	c := p.Load("gateway")
	origin, e := url.Parse(c.WebOrigin)
	if e != nil || origin.Host == "" || origin.Hostname() == c.MockHost {
		panic("Web origin must be valid and different from Mock host")
	}
	conn := p.Dial(c, c.Identity)
	defer conn.Close()
	root := os.Getenv("TESTMOON_ROOT")
	if root == "" {
		root = "."
	}
	g := &gateway{C: c, Identity: iam.NewIdentityClient(conn), Root: root, Proxies: map[string]*httputil.ReverseProxy{}}
	for prefix, svc := range map[string]string{"/api/platform/v1/": "identity", "/api/testing/v1/": "core", "/api/execution/v1/": "execution", "/api/runner/v1/": "execution", "/api/mock/v1/": "mock", "/api/share/v1/": "core"} {
		addr := os.Getenv("TESTMOON_" + strings.ToUpper(svc) + "_HTTP")
		if addr == "" {
			panic("missing HTTP address: " + svc)
		}
		g.Proxies[prefix] = proxy("http://" + addr)
	}
	addr := os.Getenv("TESTMOON_MOCK_PUBLIC_HTTP")
	if addr == "" {
		addr = "127.0.0.1:8095"
	}
	g.Mock = proxy("http://" + addr)
	server := &http.Server{Addr: c.HTTP, Handler: g, ReadHeaderTimeout: 10 * time.Second, IdleTimeout: 70 * time.Second, MaxHeaderBytes: 32 << 10}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	go func() {
		if e := server.ListenAndServe(); e != nil && !errors.Is(e, http.ErrServerClosed) {
			logrus.Error("gateway listener failed")
			stop()
		}
	}()
	logrus.WithField("address", c.HTTP).Info("gateway started")
	<-ctx.Done()
	shutdown, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = server.Shutdown(shutdown)
}
func (g *gateway) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	host := r.Host
	if h, _, e := net.SplitHostPort(host); e == nil {
		host = h
	}
	host = strings.ToLower(host)
	for k := range r.Header {
		lower := strings.ToLower(k)
		if strings.HasPrefix(lower, "x-testmoon-") || lower == "x-service-key" || lower == "request-id" {
			r.Header.Del(k)
		}
	}
	r.Header.Set("Request-Id", p.NewBase().ID)
	r.Header.Set("X-Service-Key", g.C.ServiceKey)
	if host == strings.ToLower(g.C.MockHost) {
		if !strings.HasPrefix(r.URL.Path, "/m/") {
			http.NotFound(w, r)
			return
		}
		r.Header.Del("Cookie")
		r.Header.Del("Authorization")
		g.Mock.ServeHTTP(w, r)
		return
	}
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Referrer-Policy", "same-origin")
	w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; worker-src 'self' blob:; img-src 'self' data:; frame-ancestors 'none'; base-uri 'self'; object-src 'none'")
	if r.URL.Path == "/healthz" {
		w.WriteHeader(200)
		return
	}
	if r.URL.Path == "/config.json" && r.Method == "GET" {
		origin, _ := url.Parse(g.C.WebOrigin)
		origin.Host = g.C.MockHost
		if port := origin.Port(); port != "" {
			origin.Host = net.JoinHostPort(g.C.MockHost, port)
		}
		web, _ := url.Parse(g.C.WebOrigin)
		if port := web.Port(); port != "" {
			origin.Host = net.JoinHostPort(g.C.MockHost, port)
		}
		mockOrigin := os.Getenv("TESTMOON_MOCK_ORIGIN")
		if mockOrigin == "" {
			mockOrigin = origin.String()
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-store")
		_ = json.NewEncoder(w).Encode(map[string]string{"mock_origin": mockOrigin})
		return
	}
	if strings.HasPrefix(r.URL.Path, "/api/") {
		w.Header().Set("Cache-Control", "no-store")
		if r.Method != "POST" {
			http.Error(w, "method not allowed", 405)
			return
		}
		r.Body = http.MaxBytesReader(w, r.Body, 12<<20)
		public := r.URL.Path == "/api/platform/v1/login" || r.URL.Path == "/api/share/v1/document"
		switch r.URL.Path {
		case "/api/runner/v1/register", "/api/runner/v1/heartbeat", "/api/runner/v1/poll", "/api/runner/v1/progress", "/api/runner/v1/complete":
			public = true
		}
		if !strings.HasPrefix(r.URL.Path, "/api/runner/") {
			if origin := r.Header.Get("Origin"); origin != "" && origin != g.C.WebOrigin {
				http.Error(w, "origin rejected", 403)
				return
			}
		}
		if !public {
			name := "testmoon_session"
			if g.C.SecureCookie {
				name = "__Host-testmoon_session"
			}
			cookie, e := r.Cookie(name)
			if e != nil {
				http.Error(w, "请登录", 401)
				return
			}
			ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
			reply, e := g.Identity.Session(p.Outgoing(ctx, g.C), &iam.SessionRequest{Token: cookie.Value})
			cancel()
			if e != nil {
				code := 401
				if status.Code(e) != codes.Unauthenticated {
					code = 503
				}
				http.Error(w, "登录校验失败", code)
				return
			}
			var data struct {
				Principal p.Principal `json:"principal"`
			}
			if e = p.Unpack(reply, &data); e != nil {
				http.Error(w, "登录校验失败", 503)
				return
			}
			data.Principal.SessionHash = p.Hash(cookie.Value)
			if r.URL.Path != "/api/platform/v1/session" {
				if subtle.ConstantTimeCompare([]byte(r.Header.Get("X-CSRF-Token")), []byte(data.Principal.CSRF)) != 1 {
					http.Error(w, "CSRF校验失败", 403)
					return
				}
			}
			for k, v := range p.PrincipalHeaders(g.C, data.Principal) {
				r.Header[k] = v
			}
		}
		for prefix, upstream := range g.Proxies {
			if strings.HasPrefix(r.URL.Path, prefix) {
				upstream.ServeHTTP(w, r)
				return
			}
		}
		http.NotFound(w, r)
		return
	}
	app := "portal"
	base := "/"
	if strings.HasPrefix(r.URL.Path, "/apps/testing/") || r.URL.Path == "/apps/testing" {
		app = "testing"
		base = "/apps/testing/"
	}
	relative := strings.TrimPrefix(r.URL.Path, strings.TrimSuffix(base, "/"))
	relative = path.Clean("/" + relative)
	file := filepath.Join(g.Root, "frontend", app, "dist", strings.TrimPrefix(relative, "/"))
	st, e := os.Stat(file)
	if e != nil || st.IsDir() {
		if path.Ext(relative) != "" {
			http.NotFound(w, r)
			return
		}
		file = filepath.Join(g.Root, "frontend", app, "dist", "index.html")
	}
	if _, e := os.Stat(file); e != nil {
		http.Error(w, fmt.Sprintf("%s frontend not built; run make build-all", app), 503)
		return
	}
	http.ServeFile(w, r, file)
}
