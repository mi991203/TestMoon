package api

import (
	"github.com/mi991203/TestMoon/backend/testing/core/internal/runtime"
	pb "github.com/mi991203/TestMoon/protocols/testing/core/v1"
)

type Server struct {
	pb.UnimplementedCoreServer
	deps *runtime.Dependencies
}

func New(deps *runtime.Dependencies) *Server { return &Server{deps: deps} }
