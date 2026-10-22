package main

import (
	"google.golang.org/grpc"

	handler "github.com/mi991203/TestMoon/backend/testing/execution/api"
	"github.com/mi991203/TestMoon/backend/testing/execution/conf"
	"github.com/mi991203/TestMoon/backend/testing/execution/internal/runtime"
	"github.com/mi991203/TestMoon/backend/testing/execution/logic"
	p "github.com/mi991203/TestMoon/libs/go/platform"
	pb "github.com/mi991203/TestMoon/protocols/testing/execution/v1"
)

func main() {
	c := conf.Load()
	deps := runtime.New(c)
	logic.StartQueue(deps)
	p.Serve(c, func(g *grpc.Server) { pb.RegisterExecutionServer(g, handler.New(deps)) }, pb.RegisterExecutionHandlerFromEndpoint)
}
