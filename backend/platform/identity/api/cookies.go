package api

import (
	"context"
	"net/http"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"github.com/mi991203/TestMoon/backend/platform/identity/internal/runtime"
)

func setCookie(ctx context.Context, deps *runtime.Dependencies, token string, age int) {
	name := "testmoon_session"
	if deps.Config.SecureCookie {
		name = "__Host-testmoon_session"
	}
	c := http.Cookie{Name: name, Value: token, Path: "/", HttpOnly: true, Secure: deps.Config.SecureCookie, SameSite: http.SameSiteLaxMode, MaxAge: age}
	_ = grpc.SetHeader(ctx, metadata.Pairs("set-cookie", c.String()))
}
