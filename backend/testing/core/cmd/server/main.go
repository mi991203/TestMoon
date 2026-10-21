package main

import (
	"google.golang.org/grpc"

	handler "github.com/mi991203/TestMoon/backend/testing/core/api"
	"github.com/mi991203/TestMoon/backend/testing/core/conf"
	"github.com/mi991203/TestMoon/backend/testing/core/internal/runtime"
	p "github.com/mi991203/TestMoon/libs/go/platform"
	pb "github.com/mi991203/TestMoon/protocols/testing/core/v1"
)

func main() {
	c := conf.Load()
	deps := runtime.New(c)
	defer deps.Remote.Close()
	p.Serve(c, func(g *grpc.Server) { pb.RegisterCoreServer(g, handler.New(deps)) }, pb.RegisterCoreHandlerFromEndpoint)
}
