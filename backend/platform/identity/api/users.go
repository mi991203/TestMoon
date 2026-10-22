package api

import (
	"context"
	"regexp"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	credentials "github.com/mi991203/TestMoon/backend/platform/identity/internal/password"
	"github.com/mi991203/TestMoon/backend/platform/identity/logic"
	"github.com/mi991203/TestMoon/backend/platform/identity/model"
	"github.com/mi991203/TestMoon/backend/platform/identity/serializer"
	p "github.com/mi991203/TestMoon/libs/go/platform"
	pb "github.com/mi991203/TestMoon/protocols/platform/identity/v1"
)

func (s *Server) ListUsers(ctx context.Context, r *pb.ListUsersRequest) (*pb.ListUsersResponse, error) {
	if err := validateRequest(r); err != nil {
		return nil, err
	}
	deps := s.deps
	if e := p.RequireAdmin(ctx); e != nil {
		return nil, e
	}
	var users []model.User
	e := deps.DB.Order("created_at DESC").Limit(500).Find(&users).Error
	if e != nil {
		return nil, p.DBError(e)
	}
	return &pb.ListUsersResponse{Items: serializer.Users(users)}, nil
}

func (s *Server) CreateUser(ctx context.Context, r *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {
	if err := validateRequest(r); err != nil {
		return nil, err
	}
	deps := s.deps
	if e := p.RequireAdmin(ctx); e != nil {
		return nil, e
	}
	var in struct {
		Username string `json:"username" validate:"required,max=64"`
		Name     string `json:"name" validate:"max=128"`
		Password string `json:"password"`
		Admin    bool   `json:"admin"`
	}
	if e := p.Decode(r, &in); e != nil {
		return nil, e
	}
	in.Username = strings.ToLower(in.Username)
	if !regexp.MustCompile(`^[a-z0-9][a-z0-9_.-]{2,63}$`).MatchString(in.Username) || !credentials.Valid(in.Password) {
		return nil, p.Bad("账号需为 3..64 位小写字母数字或 _. -，密码为 10..128 字节")
	}
	u := model.User{Base: p.NewBase(), Username: in.Username, Name: in.Name, Password: credentials.Hash(in.Password), Admin: in.Admin, Epoch: 1}
	if e := deps.DB.Create(&u).Error; e != nil {
		return nil, p.DBError(e)
	}
	return &pb.CreateUserResponse{User: serializer.User(u)}, nil
}

func (s *Server) ResetPassword(ctx context.Context, r *pb.ResetPasswordRequest) (*pb.ResetPasswordResponse, error) {
	if err := validateRequest(r); err != nil {
		return nil, err
	}
	deps := s.deps
	if e := p.RequireAdmin(ctx); e != nil {
		return nil, e
	}
	var in struct {
		Password string `json:"password"`
		Disabled bool   `json:"disabled"`
	}
	if e := p.Decode(r, &in); e != nil {
		return nil, e
	}
	e := deps.DB.Transaction(func(tx *gorm.DB) error {
		var users []model.User
		if e := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("admin = ?", true).Order("id").Find(&users).Error; e != nil {
			return e
		}
		var u model.User
		if e := tx.First(&u, "id = ?", r.Id).Error; e != nil {
			return e
		}
		updates := map[string]any{"epoch": gorm.Expr("epoch + 1")}

		if !credentials.Valid(in.Password) {
			return p.Bad("密码需为 10..128 字节")
		}
		updates["password"] = credentials.Hash(in.Password)

		return tx.Model(&u).Updates(updates).Error
	})
	if e != nil {
		return nil, logic.Translate(e)
	}
	return &pb.ResetPasswordResponse{Ok: true}, nil
}

func (s *Server) SetUserStatus(ctx context.Context, r *pb.SetUserStatusRequest) (*pb.SetUserStatusResponse, error) {
	if err := validateRequest(r); err != nil {
		return nil, err
	}
	deps := s.deps
	actor := p.Actor(ctx)
	if e := p.RequireAdmin(ctx); e != nil {
		return nil, e
	}
	var in struct {
		Password string `json:"password"`
		Disabled bool   `json:"disabled"`
	}
	if e := p.Decode(r, &in); e != nil {
		return nil, e
	}
	e := deps.DB.Transaction(func(tx *gorm.DB) error {
		var users []model.User
		if e := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("admin = ?", true).Order("id").Find(&users).Error; e != nil {
			return e
		}
		var u model.User
		if e := tx.First(&u, "id = ?", r.Id).Error; e != nil {
			return e
		}
		updates := map[string]any{"epoch": gorm.Expr("epoch + 1")}

		if u.ID == actor.UserID && in.Disabled {
			return p.Bad("不能禁用当前账号")
		}
		if u.Admin && in.Disabled {
			active := 0
			for _, a := range users {
				if !a.Disabled {
					active++
				}
			}
			if active <= 1 {
				return p.Bad("不能禁用最后一个平台管理员")
			}
		}
		updates["disabled"] = in.Disabled

		return tx.Model(&u).Updates(updates).Error
	})
	if e != nil {
		return nil, logic.Translate(e)
	}
	return &pb.SetUserStatusResponse{Ok: true}, nil
}
