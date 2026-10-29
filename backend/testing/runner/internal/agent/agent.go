package agent

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/mi991203/TestMoon/backend/testing/runner/internal/engine"
	"github.com/mi991203/TestMoon/libs/go/buildinfo"
	m "github.com/mi991203/TestMoon/libs/go/testing/domain"
	"github.com/mi991203/TestMoon/libs/go/transport"
	contracts "github.com/mi991203/TestMoon/protocols/testing/contracts/v1"
	pb "github.com/mi991203/TestMoon/protocols/testing/execution/v1"
	"github.com/moby/moby/api/pkg/stdcopy"
	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/client"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type State struct {
	RunnerID    string `json:"runner_id"`
	Credential  string `json:"credential"`
	Public      bool   `json:"public"`
	Concurrency int    `json:"concurrency"`
	URL         string `json:"url"`
}
type job struct {
	RunID    string     `json:"run_id"`
	Lease    string     `json:"lease"`
	Public   bool       `json:"public"`
	Snapshot m.Snapshot `json:"snapshot"`
}
type active struct {
	Lease  string
	Cancel context.CancelFunc
}
type Agent struct {
	URL, Image, StateFile string
	State                 State
	Docker                *client.Client
	HTTP                  *http.Client
	mu                    sync.Mutex
	active                map[string]active
}

func New(ctx context.Context) (*Agent, error) {
	base := strings.TrimRight(os.Getenv("TESTMOON_URL"), "/")
	u, e := url.Parse(base)
	if e != nil || u.Host == "" || u.User != nil {
		return nil, errors.New("TESTMOON_URL must be a valid origin")
	}
	if u.Scheme != "https" && (u.Scheme != "http" || os.Getenv("TESTMOON_ALLOW_HTTP") != "1") {
		return nil, errors.New("HTTPS required; local development may explicitly set TESTMOON_ALLOW_HTTP=1")
	}
	image := os.Getenv("TESTMOON_WORKER_IMAGE")
	if image == "" {
		image = "testmoon/worker:0.1.0"
	}
	stateFile := os.Getenv("TESTMOON_RUNNER_STATE")
	if stateFile == "" {
		stateFile = "/var/lib/testmoon/runner.json"
	}
	docker, e := client.New(client.FromEnv)
	if e != nil {
		return nil, e
	}
	if _, e = docker.Ping(ctx, client.PingOptions{}); e != nil {
		return nil, errors.New("Docker API unavailable")
	}
	a := &Agent{URL: base, Image: image, StateFile: stateFile, Docker: docker, HTTP: &http.Client{Timeout: 35 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return errors.New("control-plane redirects rejected") }}, active: map[string]active{}}
	if b, e := os.ReadFile(stateFile); e == nil {
		if json.Unmarshal(b, &a.State) != nil || a.State.Credential == "" || a.State.URL != base {
			return nil, errors.New("invalid Runner state or origin mismatch")
		}
	} else {
		token := os.Getenv("TESTMOON_ENROLLMENT_TOKEN")
		if token == "" {
			return nil, errors.New("first registration requires TESTMOON_ENROLLMENT_TOKEN")
		}
		response := &pb.RegisterRunnerResponse{}
		if e := a.call(ctx, "register", &pb.RegisterRunnerRequest{Token: token, Version: buildinfo.Version, Capabilities: []string{"http/1", "javascript/1"}}, response); e != nil {
			return nil, e
		}
		a.State = State{RunnerID: response.RunnerId, Credential: response.Credential, Public: response.Public, Concurrency: int(response.Concurrency), URL: base}
		if e = os.MkdirAll(filepath.Dir(stateFile), 0700); e != nil {
			return nil, e
		}
		b, _ := json.Marshal(a.State)
		if e = os.WriteFile(stateFile, b, 0600); e != nil {
			return nil, e
		}
	}
	return a, nil
}
func (a *Agent) call(ctx context.Context, op string, in, out proto.Message) error {
	payload, err := (protojson.MarshalOptions{UseProtoNames: true}).Marshal(in)
	if err != nil {
		return err
	}
	req, e := http.NewRequestWithContext(ctx, "POST", a.URL+"/api/runner/v1/"+op, bytes.NewReader(payload))
	if e != nil {
		return e
	}
	req.Header.Set("Content-Type", "application/json")
	if a.State.Credential != "" {
		req.Header.Set("Authorization", "Bearer "+a.State.Credential)
	}
	res, e := a.HTTP.Do(req)
	if e != nil {
		return fmt.Errorf("Runner %s control-plane connection failed: %w", op, e)
	}
	defer res.Body.Close()
	body, e := io.ReadAll(io.LimitReader(res.Body, 12<<20))
	if e != nil {
		return e
	}
	if res.StatusCode != 200 {
		return fmt.Errorf("Runner %s rejected (%d)", op, res.StatusCode)
	}
	if out != nil {
		return protojson.Unmarshal(body, out)
	}
	return nil
}
func (a *Agent) Run(ctx context.Context) {
	defer a.Docker.Close()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	go a.heartbeats(ctx, cancel)
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	var wg sync.WaitGroup
	defer wg.Wait()
	for {
		select {
		case <-ctx.Done():
			a.mu.Lock()
			for _, j := range a.active {
				j.Cancel()
			}
			a.mu.Unlock()
			return
		case <-ticker.C:
		}
		a.mu.Lock()
		busy := len(a.active)
		a.mu.Unlock()
		if busy >= a.State.Concurrency {
			continue
		}
		out := &pb.PollRunnerJobResponse{}
		if e := a.call(ctx, "poll", &pb.PollRunnerJobRequest{}, out); e != nil {
			logrus.Warn("Runner polling delayed")
			continue
		}
		if out.Job == nil {
			continue
		}
		var j job
		raw, _ := json.Marshal(transport.Fields(out.Job))
		if err := json.Unmarshal(raw, &j); err != nil {
			logrus.Error("invalid execution snapshot")
			continue
		}
		runCtx, runCancel := context.WithTimeout(ctx, time.Duration(j.Snapshot.TimeoutSeconds)*time.Second)
		a.mu.Lock()
		a.active[j.RunID] = active{Lease: j.Lease, Cancel: runCancel}
		a.mu.Unlock()
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer runCancel()
			result := a.execute(runCtx, j)
			completeCtx, c := context.WithTimeout(context.Background(), 10*time.Second)
			defer c()
			if e := a.report(completeCtx, "complete", j, result); e != nil {
				logrus.WithError(e).WithField("run_id", j.RunID).Warn("result delivery failed or execution already terminal")
			}
			a.mu.Lock()
			delete(a.active, j.RunID)
			a.mu.Unlock()
		}()
	}
}
func (a *Agent) heartbeats(ctx context.Context, stop context.CancelFunc) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	lastSuccess := time.Now()
	for {
		leases := map[string]string{}
		a.mu.Lock()
		for id, v := range a.active {
			leases[id] = v.Lease
		}
		a.mu.Unlock()
		out := &pb.RunnerHeartbeatResponse{}
		callCtx, cancel := context.WithTimeout(ctx, 8*time.Second)
		e := a.call(callCtx, "heartbeat", &pb.RunnerHeartbeatRequest{Leases: leases}, out)
		cancel()
		if e == nil {
			lastSuccess = time.Now()
			a.mu.Lock()
			for id, allowed := range out.Active {
				if !allowed {
					if j, ok := a.active[id]; ok {
						j.Cancel()
					}
				}
			}
			a.mu.Unlock()
		} else if time.Since(lastSuccess) > 20*time.Second {
			logrus.Error("Runner heartbeat unavailable; stopping execution")
			stop()
			return
		}
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}
func (a *Agent) execute(ctx context.Context, j job) m.Result {
	failure := func(message string) m.Result {
		state := "error"
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			state = "timed_out"
		}
		return m.Result{Status: state, Error: message}
	}
	pids := int64(128)
	created, e := a.Docker.ContainerCreate(ctx, client.ContainerCreateOptions{Config: &container.Config{Image: a.Image, User: "10001:10001", AttachStdin: true, AttachStdout: true, AttachStderr: true, OpenStdin: true, Labels: map[string]string{"testmoon.run": j.RunID}}, HostConfig: &container.HostConfig{NetworkMode: "none", ReadonlyRootfs: true, CapDrop: []string{"ALL"}, SecurityOpt: []string{"no-new-privileges:true"}, Tmpfs: map[string]string{"/tmp": "rw,nosuid,nodev,size=64m"}, AutoRemove: true, Resources: container.Resources{Memory: 512 << 20, MemorySwap: 512 << 20, NanoCPUs: 1e9, PidsLimit: &pids}}})
	if e != nil {
		return failure("隔离容器创建失败，请检查Worker镜像和Docker权限")
	}
	defer func() {
		cleanup, c := context.WithTimeout(context.Background(), 10*time.Second)
		defer c()
		_, _ = a.Docker.ContainerRemove(cleanup, created.ID, client.ContainerRemoveOptions{Force: true})
	}()
	attach, e := a.Docker.ContainerAttach(ctx, created.ID, client.ContainerAttachOptions{Stream: true, Stdin: true, Stdout: true, Stderr: true})
	if e != nil {
		return failure("无法连接隔离容器")
	}
	defer attach.Close()
	reader, writer := io.Pipe()
	defer reader.Close()
	go func() { _, e := stdcopy.StdCopy(writer, io.Discard, attach.Reader); _ = writer.CloseWithError(e) }()
	if _, e = a.Docker.ContainerStart(ctx, created.ID, client.ContainerStartOptions{}); e != nil {
		return failure("隔离容器启动失败")
	}
	encoder := json.NewEncoder(attach.Conn)
	if e = encoder.Encode(j.Snapshot); e != nil {
		return failure("无法发送执行快照")
	}
	go func() { <-ctx.Done(); attach.Close() }()
	network := &network{Public: j.Public, Clients: map[string]*http.Client{}}
	decoder := json.NewDecoder(bufio.NewReader(io.LimitReader(reader, 32<<20)))
	progress := m.Result{Status: "running"}
	for _, item := range j.Snapshot.Items {
		progress.Items = append(progress.Items, m.ItemResult{ID: item.ID, Name: item.Name, Status: "running"})
	}
	lastProgress := time.Time{}
	for {
		var frame m.Frame
		if e := decoder.Decode(&frame); e != nil {
			return failure("隔离执行进程中断")
		}
		switch frame.Type {
		case "http":
			if frame.Request == nil {
				return failure("无效执行通信")
			}
			response := network.call(ctx, frame.ItemID, *frame.Request)
			if e := encoder.Encode(m.Frame{Type: "http_result", ID: frame.ID, Response: &response}); e != nil {
				return failure("执行通信中断")
			}
		case "step":
			if frame.Step != nil {
				for i := range progress.Items {
					if progress.Items[i].ID == frame.ItemID {
						progress.Items[i].Steps = append(progress.Items[i].Steps, *frame.Step)
					}
				}
				if time.Since(lastProgress) > time.Second {
					engine.Compact(&progress)
					if err := a.report(ctx, "progress", j, progress); err != nil {
						logrus.WithError(err).WithField("run_id", j.RunID).Warn("execution progress delivery failed")
					}
					lastProgress = time.Now()
				}
			}
		case "done":
			if frame.Result == nil {
				return failure("执行进程未返回结果")
			}
			engine.Compact(frame.Result)
			return *frame.Result
		default:
			return failure("未知执行通信类型")
		}
	}
}

// report bridges the worker's IPC result into the public typed control-plane contract.
func (a *Agent) report(ctx context.Context, op string, j job, result m.Result) error {
	value := &contracts.ExecutionResult{}
	if err := transport.Encode(result, value); err != nil {
		return fmt.Errorf("encode execution %s result: %w", op, err)
	}
	if op == "progress" {
		return a.call(ctx, op, &pb.ProgressRunRequest{Id: j.RunID, Lease: j.Lease, Result: value}, &pb.ProgressRunResponse{})
	}
	return a.call(ctx, op, &pb.CompleteRunRequest{Id: j.RunID, Lease: j.Lease, Result: value}, &pb.CompleteRunResponse{})
}
