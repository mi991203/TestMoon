package thirdparty

import (
	"context"

	"google.golang.org/grpc"

	p "github.com/mi991203/TestMoon/libs/go/platform"
	core "github.com/mi991203/TestMoon/protocols/testing/core/v1"
)

type CoreClient struct {
	Config p.Config
	Conn   *grpc.ClientConn
	Client core.CoreClient
}

func NewCoreClient(c p.Config) *CoreClient {
	conn := p.Dial(c, c.Core)
	return &CoreClient{Config: c, Conn: conn, Client: core.NewCoreClient(conn)}
}

func (c *CoreClient) Authorize(ctx context.Context, org, project, permission string) error {
	_, err := c.Role(ctx, org, project, permission)
	return err
}

func (c *CoreClient) Role(ctx context.Context, org, project, permission string) (string, error) {
	out, err := c.Client.Authorize(p.Outgoing(ctx, c.Config), &core.AuthorizeRequest{OrgId: org, ProjectId: project, Permission: permission})
	if err != nil {
		return "", err
	}
	return out.Role, nil
}
