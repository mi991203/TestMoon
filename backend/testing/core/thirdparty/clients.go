// Package thirdparty owns outbound service connections; business orchestration
// remains in logic. Clients only depend on public protobuf contracts.
package thirdparty

import (
	"google.golang.org/grpc"

	p "github.com/mi991203/TestMoon/libs/go/platform"
	iam "github.com/mi991203/TestMoon/protocols/platform/identity/v1"
	execution "github.com/mi991203/TestMoon/protocols/testing/execution/v1"
	mock "github.com/mi991203/TestMoon/protocols/testing/mock/v1"
)

type Clients struct {
	Identity    iam.IdentityClient
	Execution   execution.ExecutionClient
	Mock        mock.MockClient
	connections []*grpc.ClientConn
}

func New(c p.Config) *Clients {
	identityConn := p.Dial(c, c.Identity)
	executionConn := p.Dial(c, c.Execution)
	mockConn := p.Dial(c, c.Mock)
	return &Clients{Identity: iam.NewIdentityClient(identityConn), Execution: execution.NewExecutionClient(executionConn), Mock: mock.NewMockClient(mockConn), connections: []*grpc.ClientConn{identityConn, executionConn, mockConn}}
}

func (c *Clients) Close() {
	for _, conn := range c.connections {
		_ = conn.Close()
	}
}
