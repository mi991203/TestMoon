package api

import (
	"github.com/mi991203/TestMoon/backend/testing/mock/internal/runtime"
	pb "github.com/mi991203/TestMoon/protocols/testing/mock/v1"
)

type Server struct {
	pb.UnimplementedMockServer
	deps *runtime.Dependencies
}

func New(deps *runtime.Dependencies) *Server { return &Server{deps: deps} }
