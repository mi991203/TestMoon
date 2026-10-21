package api

import (
	"github.com/mi991203/TestMoon/backend/platform/identity/internal/runtime"
	pb "github.com/mi991203/TestMoon/protocols/platform/identity/v1"
)

type Server struct {
	pb.UnimplementedIdentityServer
	deps *runtime.Dependencies
}

func New(deps *runtime.Dependencies) *Server { return &Server{deps: deps} }
