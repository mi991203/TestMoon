package main

import (
	"google.golang.org/grpc"

	handler "github.com/mi991203/TestMoon/backend/platform/identity/api"
	"github.com/mi991203/TestMoon/backend/platform/identity/conf"
	"github.com/mi991203/TestMoon/backend/platform/identity/internal/runtime"
	"github.com/mi991203/TestMoon/backend/platform/identity/logic"
	p "github.com/mi991203/TestMoon/libs/go/platform"
	pb "github.com/mi991203/TestMoon/protocols/platform/identity/v1"
)

func main() {
	c := conf.Load()
	deps := runtime.New(c)
	logic.Bootstrap(deps)
	p.Serve(c, func(g *grpc.Server) { pb.RegisterIdentityServer(g, handler.New(deps)) }, pb.RegisterIdentityHandlerFromEndpoint)
}
