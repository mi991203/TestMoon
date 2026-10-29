package main

import (
	"context"
	"encoding/json"
	"github.com/mi991203/TestMoon/backend/testing/runner/internal/agent"
	"github.com/mi991203/TestMoon/libs/go/buildinfo"
	"github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	if len(os.Args) == 2 && (os.Args[1] == "--version" || os.Args[1] == "version") {
		_ = json.NewEncoder(os.Stdout).Encode(buildinfo.Current())
		return
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	a, e := agent.New(ctx)
	if e != nil {
		logrus.Fatal(e)
	}
	a.Run(ctx)
}
