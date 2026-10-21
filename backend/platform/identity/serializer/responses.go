// Package serializer exposes only explicitly selected public account fields.
package serializer

import (
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/mi991203/TestMoon/backend/platform/identity/model"
	p "github.com/mi991203/TestMoon/libs/go/platform"
	pb "github.com/mi991203/TestMoon/protocols/platform/identity/v1"
)

func User(value model.User) *pb.User {
	return &pb.User{Id: value.ID, Username: value.Username, Name: value.Name, Admin: value.Admin, Disabled: value.Disabled, CreatedAt: timestamppb.New(value.CreatedAt), UpdatedAt: timestamppb.New(value.UpdatedAt)}
}

func Users(values []model.User) []*pb.User {
	items := make([]*pb.User, 0, len(values))
	for _, value := range values {
		items = append(items, User(value))
	}
	return items
}

func Organization(value model.Organization) *pb.Organization {
	return &pb.Organization{Id: value.ID, Name: value.Name, CreatedAt: timestamppb.New(value.CreatedAt), UpdatedAt: timestamppb.New(value.UpdatedAt)}
}

func Organizations(values []model.Organization) []*pb.Organization {
	items := make([]*pb.Organization, 0, len(values))
	for _, value := range values {
		items = append(items, Organization(value))
	}
	return items
}

func Principal(value p.Principal) *pb.Principal {
	return &pb.Principal{UserId: value.UserID, Username: value.Username, Admin: value.Admin, Organizations: value.Organizations, Csrf: value.CSRF}
}

func Member(member model.Member, user model.User) *pb.OrganizationMember {
	return &pb.OrganizationMember{Id: member.ID, OrgId: member.OrgID, UserId: user.ID, Username: user.Username, Name: user.Name, Role: member.Role, Disabled: user.Disabled}
}
