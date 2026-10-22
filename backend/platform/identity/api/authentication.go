package api

import (
	"context"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"

	credentials "github.com/mi991203/TestMoon/backend/platform/identity/internal/password"
	"github.com/mi991203/TestMoon/backend/platform/identity/logic"
	"github.com/mi991203/TestMoon/backend/platform/identity/model"
	"github.com/mi991203/TestMoon/backend/platform/identity/serializer"
	p "github.com/mi991203/TestMoon/libs/go/platform"
	pb "github.com/mi991203/TestMoon/protocols/platform/identity/v1"
)

func (s *Server) Login(ctx context.Context, r *pb.LoginRequest) (*pb.LoginResponse, error) {
	if err := validateRequest(r); err != nil {
		return nil, err
	}
	deps := s.deps
	var in struct {
		Username string `json:"username" validate:"required,max=64"`
		Password string `json:"password" validate:"required,max=128"`
	}
	if e := p.Decode(r, &in); e != nil {
		return nil, e
	}
	in.Username = strings.ToLower(in.Username)
	result, err := logic.Authenticate(ctx, deps, in.Username, in.Password)
	if err != nil {
		return nil, err
	}
	setCookie(ctx, deps, result.Token, 8*3600)
	return &pb.LoginResponse{Csrf: result.CSRF, User: serializer.User(result.User)}, nil
}

func (s *Server) Logout(ctx context.Context, r *pb.LogoutRequest) (*pb.LogoutResponse, error) {
	if err := validateRequest(r); err != nil {
		return nil, err
	}
	deps := s.deps
	actor := p.Actor(ctx)
	if actor.SessionHash != "" {
		if e := deps.Redis.Del(ctx, "session:"+actor.SessionHash).Err(); e != nil {
			return nil, status.Error(codes.Unavailable, "退出失败")
		}
	}
	setCookie(ctx, deps, "", -1)
	return &pb.LogoutResponse{Ok: true}, nil
}

func (s *Server) Session(ctx context.Context, r *pb.SessionRequest) (*pb.SessionResponse, error) {
	if err := validateRequest(r); err != nil {
		return nil, err
	}
	deps := s.deps
	actor := p.Actor(ctx)
	token := r.Token
	principal, e := logic.ResolvePrincipal(ctx, deps, token, actor.SessionHash)
	if e != nil {
		return nil, e
	}
	var orgs []model.Organization
	ids := []string{}
	for id := range principal.Organizations {
		ids = append(ids, id)
	}
	if len(ids) > 0 {
		deps.DB.Where("id IN ?", ids).Order("name,id").Find(&orgs)
	}
	return &pb.SessionResponse{Principal: serializer.Principal(principal), Organizations: serializer.Organizations(orgs), Csrf: principal.CSRF}, nil
}

func (s *Server) ChangePassword(ctx context.Context, r *pb.ChangePasswordRequest) (*pb.ChangePasswordResponse, error) {
	if err := validateRequest(r); err != nil {
		return nil, err
	}
	deps := s.deps
	actor := p.Actor(ctx)
	if actor.UserID == "" {
		return nil, p.Denied()
	}
	var in struct {
		Old string `json:"old_password"`
		New string `json:"new_password"`
	}
	if e := p.Decode(r, &in); e != nil {
		return nil, e
	}
	if !credentials.Valid(in.New) {
		return nil, p.Bad("密码需为 10..128 字节")
	}
	var u model.User
	if e := deps.DB.First(&u, "id = ?", actor.UserID).Error; e != nil {
		return nil, p.DBError(e)
	}
	if !credentials.Verify(u.Password, in.Old) {
		return nil, p.Bad("原密码错误")
	}
	e := deps.DB.Model(&u).Updates(map[string]any{"password": credentials.Hash(in.New), "epoch": gorm.Expr("epoch + 1")}).Error
	if e != nil {
		return nil, p.DBError(e)
	}
	setCookie(ctx, deps, "", -1)
	return &pb.ChangePasswordResponse{Ok: true}, nil
}
