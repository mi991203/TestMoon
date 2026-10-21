package logic

import (
	"context"

	"github.com/mi991203/TestMoon/backend/testing/core/internal/runtime"
	p "github.com/mi991203/TestMoon/libs/go/platform"
	iam "github.com/mi991203/TestMoon/protocols/platform/identity/v1"
)

func VerifyOrgMember(ctx context.Context, deps *runtime.Dependencies, org, user string) error {
	out, e := deps.Remote.Identity.ListOrganizationMembers(p.Outgoing(ctx, deps.Config), &iam.ListOrganizationMembersRequest{OrgId: org})
	if e != nil {
		return e
	}
	for _, member := range out.Items {
		if member.UserId == user && !member.Disabled {
			return nil
		}
	}
	return p.Bad("请选择本组织有效成员")
}
