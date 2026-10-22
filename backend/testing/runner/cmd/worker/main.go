package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/mi991203/TestMoon/backend/testing/runner/internal/engine"
	m "github.com/mi991203/TestMoon/libs/go/testing/domain"
	"os"
	"time"
)

func main() {
	encoder := json.NewEncoder(os.Stdout)
	decoder := json.NewDecoder(os.Stdin)
	var snapshot m.Snapshot
	if decoder.Decode(&snapshot) != nil || snapshot.Version != 1 {
		os.Exit(2)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(snapshot.TimeoutSeconds)*time.Second)
	defer cancel()
	sequence := 0
	e := engine.Engine{Event: func(item string, step m.StepResult) {
		_ = encoder.Encode(m.Frame{Type: "step", ItemID: item, Step: &step})
	}}
	e.Transport = func(ctx context.Context, item string, r m.HTTPRequest) m.Response {
		sequence++
		id := fmt.Sprint(sequence)
		if encoder.Encode(m.Frame{Type: "http", ID: id, ItemID: item, Request: &r}) != nil {
			return m.Response{Error: "执行通信失败"}
		}
		var frame m.Frame
		if decoder.Decode(&frame) != nil || frame.Type != "http_result" || frame.ID != id || frame.Response == nil {
			return m.Response{Error: "执行通信失败"}
		}
		return *frame.Response
	}
	result := e.Run(ctx, snapshot)
	_ = encoder.Encode(m.Frame{Type: "done", Result: &result})
}
