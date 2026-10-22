package main

import (
	"net/http"
	"time"

	"google.golang.org/grpc"

	handler "github.com/mi991203/TestMoon/backend/testing/mock/api"
	"github.com/mi991203/TestMoon/backend/testing/mock/conf"
	"github.com/mi991203/TestMoon/backend/testing/mock/internal/runtime"
	p "github.com/mi991203/TestMoon/libs/go/platform"
	pb "github.com/mi991203/TestMoon/protocols/testing/mock/v1"
)

func main() {
	c := conf.Load()
	deps := runtime.New(c)
	public := &http.Server{Addr: conf.PublicHTTPAddress(), Handler: handler.NewPublicHandler(deps), ReadHeaderTimeout: 10 * time.Second, IdleTimeout: 30 * time.Second}
	defer public.Close()
	go func() {
		if err := public.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			panic(err)
		}
	}()
	p.Serve(c, func(g *grpc.Server) { pb.RegisterMockServer(g, handler.New(deps)) }, pb.RegisterMockHandlerFromEndpoint)
}
