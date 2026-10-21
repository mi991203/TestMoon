package logic

import (
	"context"
	"encoding/json"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	credentials "github.com/mi991203/TestMoon/backend/platform/identity/internal/password"
	"github.com/mi991203/TestMoon/backend/platform/identity/internal/runtime"
	"github.com/mi991203/TestMoon/backend/platform/identity/model"
	p "github.com/mi991203/TestMoon/libs/go/platform"
)

type LoginSession struct {
	User        model.User
	Token, CSRF string
}

func Authenticate(ctx context.Context, deps *runtime.Dependencies, username, password string) (LoginSession, error) {
	key := "login:" + p.Hash(username)
	n, e := deps.Redis.Incr(ctx, key).Result()
	if e != nil {
		return LoginSession{}, status.Error(codes.Unavailable, "会话存储不可用")
	}
	if n == 1 {
		_ = deps.Redis.Expire(ctx, key, 15*time.Minute).Err()
	}
	if n > 10 {
		return LoginSession{}, status.Error(codes.ResourceExhausted, "登录尝试过多，请稍后重试")
	}
	var u model.User
	err := deps.DB.First(&u, "username = ?", username).Error
	if err != nil || u.Disabled || !credentials.Verify(u.Password, password) {
		return LoginSession{}, status.Error(codes.Unauthenticated, "账号或密码错误")
	}
	token := p.Token()
	v := sessionRecord{UserID: u.ID, Epoch: u.Epoch, CSRF: p.Token()}
	b, _ := json.Marshal(v)
	if e = deps.Redis.Set(ctx, "session:"+p.Hash(token), b, 8*time.Hour).Err(); e != nil {
		return LoginSession{}, status.Error(codes.Unavailable, "会话存储不可用")
	}
	_ = deps.Redis.Del(ctx, key).Err()
	return LoginSession{User: u, Token: token, CSRF: v.CSRF}, nil
}
