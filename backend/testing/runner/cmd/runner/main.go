package main

import (
	"context"
	"github.com/mi991203/TestMoon/backend/testing/runner/internal/agent"
	"github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	a, e := agent.New(ctx)
	if e != nil {
		logrus.Fatal(e)
	}
	a.Run(ctx)
}
