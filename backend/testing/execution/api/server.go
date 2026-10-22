package api

import (
	"github.com/mi991203/TestMoon/backend/testing/execution/internal/runtime"
	pb "github.com/mi991203/TestMoon/protocols/testing/execution/v1"
)

type Server struct {
	pb.UnimplementedExecutionServer
	deps *runtime.Dependencies
}

func New(deps *runtime.Dependencies) *Server { return &Server{deps: deps} }
