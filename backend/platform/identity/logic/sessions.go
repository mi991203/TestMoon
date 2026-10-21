package logic

import (
	"context"
	"encoding/json"
	"os"
	"regexp"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	credentials "github.com/mi991203/TestMoon/backend/platform/identity/internal/password"
	"github.com/mi991203/TestMoon/backend/platform/identity/internal/runtime"
	"github.com/mi991203/TestMoon/backend/platform/identity/model"
	p "github.com/mi991203/TestMoon/libs/go/platform"
)

type sessionRecord struct {
	UserID string `json:"user_id"`
	Epoch  int64  `json:"epoch"`
	CSRF   string `json:"csrf"`
}

func Bootstrap(deps *runtime.Dependencies) {
	username := strings.ToLower(os.Getenv("TESTMOON_BOOTSTRAP_USERNAME"))
	password := os.Getenv("TESTMOON_BOOTSTRAP_PASSWORD")
	var n int64
	if e := deps.DB.Model(&model.User{}).Count(&n).Error; e != nil {
		panic(e)
	}
	if n != 0 {
		return
	}
	if !regexp.MustCompile(`^[a-z0-9][a-z0-9_.-]{2,63}$`).MatchString(username) || !credentials.Valid(password) {
		panic("empty identity database requires TESTMOON_BOOTSTRAP_USERNAME and a password of 10..128 bytes")
	}
	if e := deps.DB.Create(&model.User{Base: p.NewBase(), Username: username, Name: username, Password: credentials.Hash(password), Admin: true, Epoch: 1}).Error; e != nil {
		panic(e)
	}
}

func ResolvePrincipal(ctx context.Context, deps *runtime.Dependencies, token, hash string) (p.Principal, error) {
	if token != "" {
		hash = p.Hash(token)
	}
	if hash == "" {
		return p.Principal{}, status.Error(codes.Unauthenticated, "请登录")
	}
	raw, e := deps.Redis.Get(ctx, "session:"+hash).Result()
	if e != nil {
		return p.Principal{}, status.Error(codes.Unauthenticated, "登录已失效")
	}
	var v sessionRecord
	if json.Unmarshal([]byte(raw), &v) != nil {
		return p.Principal{}, status.Error(codes.Unauthenticated, "登录已失效")
	}
	var u model.User
	if deps.DB.First(&u, "id = ?", v.UserID).Error != nil || u.Disabled || u.Epoch != v.Epoch {
		_ = deps.Redis.Del(ctx, "session:"+hash).Err()
		return p.Principal{}, status.Error(codes.Unauthenticated, "登录已失效")
	}
	principal := p.Principal{UserID: u.ID, Username: u.Username, Admin: u.Admin, SessionHash: hash, CSRF: v.CSRF, Organizations: map[string]string{}}
	var members []model.Member
	if e = deps.DB.Where("user_id = ?", u.ID).Find(&members).Error; e != nil {
		return principal, p.DBError(e)
	}
	for _, m := range members {
		var o model.Organization
		if deps.DB.First(&o, "id = ?", m.OrgID).Error == nil {
			principal.Organizations[m.OrgID] = m.Role
		}
	}
	return principal, nil
}
