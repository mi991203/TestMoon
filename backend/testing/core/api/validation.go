package api

import (
	"google.golang.org/protobuf/proto"

	p "github.com/mi991203/TestMoon/libs/go/platform"
)

// validateRequest checks transport-level presence. Permissions and resource
// relationships are resolved after the request has a valid basic shape.
func validateRequest(request proto.Message) error {
	if request == nil || !request.ProtoReflect().IsValid() {
		return p.Bad("请求不能为空")
	}
	if scoped, ok := request.(interface{ GetOrgId() string }); ok && scoped.GetOrgId() == "" {
		return p.Bad("缺少组织ID")
	}
	if scoped, ok := request.(interface{ GetProjectId() string }); ok && scoped.GetProjectId() == "" {
		return p.Bad("缺少项目ID")
	}
	if identified, ok := request.(interface{ GetId() string }); ok && identified.GetId() == "" {
		return p.Bad("缺少资源ID")
	}
	if member, ok := request.(interface{ GetUserId() string }); ok && member.GetUserId() == "" {
		return p.Bad("缺少成员ID")
	}
	return nil
}
