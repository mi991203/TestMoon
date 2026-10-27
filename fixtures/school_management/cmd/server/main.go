package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/mi991203/TestMoon/fixtures/school_management/api"
	"github.com/mi991203/TestMoon/fixtures/school_management/conf"
	"github.com/mi991203/TestMoon/fixtures/school_management/dao"
	"github.com/mi991203/TestMoon/fixtures/school_management/logic"
	pb "github.com/mi991203/TestMoon/protocols/fixtures/school_management/v1"
)

func main() {
	logrus.SetFormatter(&logrus.JSONFormatter{})
	if err := run(); err != nil {
		logrus.Error(err)
		os.Exit(1)
	}
}

func run() error {
	c, err := conf.Load()
	if err != nil {
		return err
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	startup, cancel := context.WithTimeout(ctx, 30*time.Second)
	db, err := dao.Open(startup, c.DSN)
	cancel()
	if err != nil {
		return err
	}
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	defer sqlDB.Close()

	// Bind both listeners before serving, so a port conflict fails startup immediately.
	grpcListener, err := net.Listen("tcp", c.GRPC)
	if err != nil {
		return fmt.Errorf("gRPC listen: %w", err)
	}
	defer grpcListener.Close()
	httpListener, err := net.Listen("tcp", c.HTTP)
	if err != nil {
		return fmt.Errorf("HTTP listen: %w", err)
	}
	defer httpListener.Close()
	server := grpc.NewServer(grpc.UnaryInterceptor(func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		id := uuid.NewString()
		defer func() {
			if recover() != nil {
				logrus.WithFields(logrus.Fields{"request_id": id, "method": info.FullMethod}).Error("request panic")
				resp, err = nil, status.Error(codes.Internal, "service internal error")
			}
		}()
		return handler(ctx, req)
	}))
	defer server.Stop()
	pb.RegisterSchoolManagementServer(server, &api.Server{Service: &logic.Service{DB: db}})
	reflection.Register(server)

	// Dial loopback when the server binds a wildcard address.
	host, port, err := net.SplitHostPort(grpcListener.Addr().String())
	if err != nil {
		return err
	}
	if host == "" || host == "0.0.0.0" {
		host = "127.0.0.1"
	}
	if host == "::" {
		host = "::1"
	}
	conn, err := grpc.NewClient(net.JoinHostPort(host, port), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("gateway connection failed")
	}
	defer conn.Close()
	mux := runtime.NewServeMux(runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{
		MarshalOptions:   protojson.MarshalOptions{UseProtoNames: true, EmitUnpopulated: true},
		UnmarshalOptions: protojson.UnmarshalOptions{DiscardUnknown: false},
	}))
	if err = pb.RegisterSchoolManagementHandler(ctx, mux, conn); err != nil {
		return err
	}
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/healthz" {
			checkCtx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
			defer cancel()
			if sqlDB.PingContext(checkCtx) != nil {
				http.Error(w, "database unavailable", http.StatusServiceUnavailable)
				return
			}
			w.WriteHeader(http.StatusOK)
			return
		}
		r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
		mux.ServeHTTP(w, r)
	})
	httpServer := &http.Server{Handler: handler, ReadHeaderTimeout: 10 * time.Second, ReadTimeout: 30 * time.Second, WriteTimeout: 30 * time.Second, IdleTimeout: 65 * time.Second}
	defer httpServer.Close()
	errorsCh := make(chan error, 2)
	go func() { errorsCh <- server.Serve(grpcListener) }()
	go func() { errorsCh <- httpServer.Serve(httpListener) }()
	logrus.WithFields(logrus.Fields{"http": httpListener.Addr().String(), "grpc": grpcListener.Addr().String()}).Info("school_management started")
	var serveErr error
	select {
	case <-ctx.Done():
	case serveErr = <-errorsCh:
		stop()
	}
	shutdown, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = httpServer.Shutdown(shutdown)
	done := make(chan struct{})
	go func() { server.GracefulStop(); close(done) }()
	select {
	case <-done:
	case <-shutdown.Done():
		server.Stop()
	}
	if serveErr != nil && !errors.Is(serveErr, http.ErrServerClosed) && !errors.Is(serveErr, grpc.ErrServerStopped) {
		return serveErr
	}
	return nil
}
