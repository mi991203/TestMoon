package apimoon_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/mi991203/TestMoon/tools/apimoon/cli"
	"github.com/mi991203/TestMoon/tools/apimoon/clierror"
	"github.com/mi991203/TestMoon/tools/apimoon/config"
	"github.com/mi991203/TestMoon/tools/apimoon/jsonutil"
	"github.com/mi991203/TestMoon/tools/apimoon/plan"
	"github.com/mi991203/TestMoon/tools/apimoon/platform"
)

type testPlatform struct {
	mu               sync.Mutex
	assets           map[string]map[string]any
	sequence, writes int
	dropKind, deny   string
	runState         string
}

func fixture(t *testing.T) (*testPlatform, *platform.Client, string) {
	t.Helper()
	f := &testPlatform{assets: map[string]map[string]any{}, runState: "passed"}
	server := httptest.NewServer(http.HandlerFunc(f.serve))
	t.Cleanup(server.Close)
	c := config.Config{BaseURL: server.URL, HTTPTimeoutSeconds: 2, Auth: config.AuthConfig{BaseURL: server.URL, CookieName: "testmoon_session", CookieValue: "private-session", CSRF: "private-csrf", UserID: "user"}, Context: config.Scope{OrgID: "org", ProjectID: "project", RunnerID: "public"}}
	file := filepath.Join(t.TempDir(), "config.yaml")
	if err := config.Save(file, c); err != nil {
		t.Fatal(err)
	}
	return f, platform.New(c), file
}
func (f *testPlatform) serve(w http.ResponseWriter, r *http.Request) {
	f.mu.Lock()
	defer f.mu.Unlock()
	w.Header().Set("Content-Type", "application/json")
	send := func(v any) { _ = json.NewEncoder(w).Encode(v) }
	var in map[string]any
	_ = json.NewDecoder(r.Body).Decode(&in)
	if strings.HasSuffix(r.URL.Path, "/login") {
		http.SetCookie(w, &http.Cookie{Name: "testmoon_session", Value: "private-session", MaxAge: 28800, Path: "/"})
		send(map[string]any{"csrf": "private-csrf", "user": map[string]any{"id": "user", "username": "tester"}})
		return
	}
	cookie, err := r.Cookie("testmoon_session")
	if err != nil || cookie.Value != "private-session" {
		w.WriteHeader(401)
		send(map[string]any{"message": "请登录"})
		return
	}
	if !strings.HasSuffix(r.URL.Path, "/session") && r.Header.Get("X-CSRF-Token") != "private-csrf" {
		w.WriteHeader(403)
		send(map[string]any{"message": "CSRF校验失败"})
		return
	}
	if strings.HasSuffix(r.URL.Path, "/session") {
		send(map[string]any{"csrf": "private-csrf", "principal": map[string]any{"user_id": "user", "csrf": "private-csrf"}})
		return
	}
	if strings.HasSuffix(r.URL.Path, "/logout") {
		send(map[string]any{"ok": true})
		return
	}
	if strings.HasSuffix(r.URL.Path, "/authorize") {
		send(map[string]any{"role": "editor", "active": true})
		return
	}
	if strings.HasSuffix(r.URL.Path, "/project-runners/list") {
		send(map[string]any{"items": []any{map[string]any{"id": "public", "online": true, "public": true}}})
		return
	}
	if strings.HasSuffix(r.URL.Path, "/run-task") {
		send(map[string]any{"run": map[string]any{"id": "run", "state": "queued", "task_id": in["task_id"]}})
		return
	}
	if strings.HasSuffix(r.URL.Path, "/run") {
		send(map[string]any{"run": map[string]any{"id": "run", "state": f.runState}, "result": map[string]any{"status": f.runState, "items": []any{map[string]any{"id": "item", "steps": []any{map[string]any{"id": "step", "status": f.runState}}}}}})
		return
	}
	if strings.HasSuffix(r.URL.Path, "/cancel") {
		send(map[string]any{"ok": true})
		return
	}
	segments := strings.Split(r.URL.Path, "/")
	last := segments[len(segments)-1]
	kind := ""
	action := last
	for k, plural := range map[string]string{"api": "apis", "case": "cases", "scenario": "scenarios", "environment": "environments", "folder": "folders"} {
		if strings.Contains(r.URL.Path, "/"+plural+"/") {
			kind = k
		}
	}
	if strings.Contains(r.URL.Path, "/execution/") {
		kind = "task"
		action = map[string]string{"tasks": "list", "create-task": "create", "update-task": "update", "delete-task": "delete"}[last]
	}
	if f.deny == kind+"/"+action {
		w.WriteHeader(403)
		send(map[string]any{"message": "没有权限"})
		return
	}
	if kind == "" {
		w.WriteHeader(404)
		send(map[string]any{"message": "不存在"})
		return
	}
	switch action {
	case "list":
		items := []any{}
		for _, v := range f.assets {
			if v["kind"] == kind {
				copy := jsonutil.Clone(v)
				delete(copy, "kind")
				items = append(items, copy)
			}
		}
		send(map[string]any{"items": items})
	case "get":
		v, ok := f.assets[jsonutil.String(in["id"])]
		if !ok || v["kind"] != kind {
			w.WriteHeader(404)
			send(map[string]any{"message": "不存在"})
			return
		}
		copy := jsonutil.Clone(v)
		delete(copy, "kind")
		send(map[string]any{"resource": copy})
	case "create", "update":
		id := jsonutil.String(in["id"])
		revision := 1
		if action == "create" {
			f.sequence++
			id = fmt.Sprintf("asset-%d", f.sequence)
		} else {
			old, ok := f.assets[id]
			if !ok {
				w.WriteHeader(404)
				send(map[string]any{"message": "不存在"})
				return
			}
			if jsonutil.String(old["revision"]) != jsonutil.String(in["revision"]) {
				w.WriteHeader(409)
				send(map[string]any{"message": "版本冲突"})
				return
			}
			revision = int(old["revision"].(int)) + 1
		}
		v := map[string]any{"id": id, "kind": kind, "name": in["name"], "revision": revision, "document": in["document"]}
		if folder := jsonutil.String(in["folder_id"]); folder != "" {
			v["folder_id"] = folder
		}
		f.assets[id] = v
		f.writes++
		if f.dropKind == kind {
			f.dropKind = ""
			conn, _, _ := w.(http.Hijacker).Hijack()
			_ = conn.Close()
			return
		}
		if kind == "task" {
			send(map[string]any{"id": id})
			return
		}
		copy := jsonutil.Clone(v)
		delete(copy, "kind")
		send(map[string]any{"resource": copy})
	case "delete":
		delete(f.assets, jsonutil.String(in["id"]))
		send(map[string]any{"ok": true})
	default:
		w.WriteHeader(400)
		send(map[string]any{"message": "未知操作"})
	}
}
func samplePlan() plan.Plan {
	return plan.Plan{Version: 1, PlanID: "school-v1", Entries: []plan.Entry{
		{Key: "task", Kind: "task", Name: "school", Document: map[string]any{"items": []any{map[string]any{"kind": "folder", "name": "group", "children": []any{map[string]any{"kind": "scenario", "id": "ref:scenario", "node_id": "scenario-instance"}}}}}},
		{Key: "scenario", Kind: "scenario", Name: "school", Document: map[string]any{"steps": []any{map[string]any{"id": "step", "case_id": "ref:case"}}}},
		{Key: "case", Kind: "case", Name: "school", Document: map[string]any{"api_id": "ref:api", "assertions": []any{map[string]any{"source": "status", "operator": "eq", "expected": 200}}, "before": []any{map[string]any{"type": "javascript", "version": 1, "script": "console.log('ref:api')"}}}},
		{Key: "api", Kind: "api", Name: "school", Document: map[string]any{"protocol": "http", "request": map[string]any{"method": "GET", "url": "https://example.test/schools", "body": "ref:case"}}},
	}}
}
func TestLoginAndConfigSecurity(t *testing.T) {
	_, _, file := fixture(t)
	var out, logs bytes.Buffer
	code := cli.Execute(context.Background(), []string{"auth", "login", "--config", file, "--username", "tester", "--password-stdin"}, strings.NewReader("long-password\n"), &out, &logs)
	if code != 0 {
		t.Fatalf("login: %s", out.String())
	}
	if strings.Contains(out.String(), "private-") {
		t.Fatal("credentials leaked")
	}
	c, err := config.Load(file)
	if err != nil || c.Auth.CookieValue != "private-session" || c.Auth.UserID != "user" {
		t.Fatal(c, err)
	}
	info, _ := os.Stat(file)
	if info.Mode().Perm() != 0600 {
		t.Fatal("unsafe config permissions")
	}
	out.Reset()
	if cli.Execute(context.Background(), []string{"--config", file, "auth", "status"}, nil, &out, &logs) != 0 {
		t.Fatal(out.String())
	}
	if strings.Contains(out.String(), "private-csrf") {
		t.Fatal("CSRF leaked in status")
	}
	out.Reset()
	if cli.Execute(context.Background(), []string{"--config", file, "config", "set", "base_url", "https://other.test"}, nil, &out, &logs) != 0 {
		t.Fatal(out.String())
	}
	c, _ = config.Load(file)
	if c.Auth.CookieValue != "" {
		t.Fatal("credentials survived endpoint change")
	}
}
func TestSessionCannotFollowRedirectOrCrossServer(t *testing.T) {
	_, c, _ := fixture(t)
	called := false
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { called = true }))
	defer target.Close()
	origin := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { http.Redirect(w, r, target.URL, 307) }))
	defer origin.Close()
	c.Config.BaseURL = origin.URL
	c.Config.Auth.BaseURL = origin.URL
	_, err := c.Call(context.Background(), "identity", "Session", map[string]any{})
	if err == nil || called {
		t.Fatal("redirect followed")
	}
	c.Config.BaseURL = target.URL
	_, err = c.Call(context.Background(), "identity", "Session", map[string]any{})
	if clierror.From(err).Code != "auth" || called {
		t.Fatal("cross-server credential reuse")
	}
}
func TestPlanSyncReferencesReplayUpdateAndConflict(t *testing.T) {
	f, c, file := fixture(t)
	s := plan.Syncer{Client: c, ConfigFile: file}
	p := samplePlan()
	state, err := s.Sync(context.Background(), p)
	if err != nil {
		t.Fatal(err)
	}
	if len(state.Entries) != 4 || f.writeCount() != 4 {
		t.Fatal(state, f.writeCount())
	}
	caseID := state.Entries["case"].ID
	apiID := state.Entries["api"].ID
	if jsonutil.Object(f.asset(caseID)["document"])["api_id"] != apiID {
		t.Fatal("case API reference unresolved")
	}
	if jsonutil.Object(jsonutil.Object(f.asset(apiID)["document"])["request"])["body"] != "ref:case" {
		t.Fatal("literal payload replaced")
	}
	if _, err = s.Sync(context.Background(), p); err != nil {
		t.Fatal(err)
	}
	if f.writeCount() != 4 {
		t.Fatal("replay wrote resources", f.writeCount())
	}
	p.Entries[3].Name = "changed"
	if _, err = s.Sync(context.Background(), p); err != nil {
		t.Fatal(err)
	}
	if f.writeCount() != 5 {
		t.Fatal("unexpected update count", f.writeCount())
	}
	f.mu.Lock()
	f.assets[caseID]["revision"] = 2
	f.mu.Unlock()
	if _, err = s.Sync(context.Background(), p); err == nil || clierror.From(err).Code != "conflict" {
		t.Fatal("human edit not detected", err)
	}
}
func TestUncertainCreateRequiresVerifiedBinding(t *testing.T) {
	f, c, file := fixture(t)
	f.dropKind = "api"
	s := plan.Syncer{Client: c, ConfigFile: file}
	p := samplePlan()
	state, err := s.Sync(context.Background(), p)
	if err == nil || clierror.From(err).Code != "uncertain" {
		t.Fatal(state, err)
	}
	if _, err = s.Sync(context.Background(), p); clierror.From(err).Code != "uncertain" || f.writeCount() != 1 {
		t.Fatal("uncertain create retried")
	}
	if _, err = s.Bind(context.Background(), p, "api", "asset-1"); err != nil {
		t.Fatal(err)
	}
	if _, err = s.Sync(context.Background(), p); err != nil {
		t.Fatal(err)
	}
	if f.writeCount() != 4 {
		t.Fatal("duplicated create", f.writeCount())
	}
}
func TestPlanValidationRejectsBadGraphWithoutWrites(t *testing.T) {
	f, c, _ := fixture(t)
	for _, mutate := range []func(*plan.Plan){
		func(p *plan.Plan) { p.Entries[2].Document["api_id"] = "ref:missing" },
		func(p *plan.Plan) { p.Entries[2].Document["api_id"] = "ref:scenario" },
		func(p *plan.Plan) { p.Entries[3].Document["unknown"] = true },
		func(p *plan.Plan) { p.Entries[0].Document["items"] = []any{} },
		func(p *plan.Plan) { p.Entries[0].Document["timeout_seconds"] = "-1" },
		func(p *plan.Plan) {
			p.Entries = append(p.Entries, plan.Entry{Key: "f1", Kind: "folder", Name: "f1", FolderID: "ref:f2", Document: map[string]any{"kind": "api"}}, plan.Entry{Key: "f2", Kind: "folder", Name: "f2", FolderID: "ref:f1", Document: map[string]any{"kind": "api"}})
		},
	} {
		p := samplePlan()
		mutate(&p)
		if _, err := plan.Validate(context.Background(), c, p); err == nil {
			t.Fatal("invalid plan accepted")
		}
	}
	if f.writeCount() != 0 {
		t.Fatal("validation wrote assets")
	}
}
func TestTaskPostWriteReadFailureIsUncertain(t *testing.T) {
	f, c, file := fixture(t)
	f.deny = "task/list"
	s := plan.Syncer{Client: c, ConfigFile: file}
	p := samplePlan()
	_, err := s.Sync(context.Background(), p)
	if err == nil || clierror.From(err).Code != "uncertain" {
		t.Fatal(err)
	}
	state, _ := s.Status(p.PlanID)
	if state.Entries["task"].State != "pending" {
		t.Fatal("forgot successful write")
	}
	f.mu.Lock()
	f.deny = ""
	id := ""
	for k, v := range f.assets {
		if v["kind"] == "task" {
			id = k
		}
	}
	f.mu.Unlock()
	if _, err = s.Bind(context.Background(), p, "task", id); err != nil {
		t.Fatal(err)
	}
}
func TestRunWaitAndQueryExitCodes(t *testing.T) {
	f, _, file := fixture(t)
	var out, logs bytes.Buffer
	call := func(args ...string) int {
		out.Reset()
		return cli.Execute(context.Background(), append([]string{"--config", file}, args...), nil, &out, &logs)
	}
	if code := call("task", "run", "--id", "task", "--wait"); code != 0 {
		t.Fatal(code, out.String())
	}
	if !strings.Contains(out.String(), "idempotency_key") || !strings.Contains(out.String(), "reports/run") {
		t.Fatal(out.String())
	}
	f.setRunState("failed")
	if code := call("run", "get", "--id", "run"); code != 0 {
		t.Fatal("report query failed", code)
	}
	if code := call("run", "wait", "--id", "run"); code != 6 {
		t.Fatal(code, out.String())
	}
	f.setRunState("unverified")
	if code := call("run", "wait", "--id", "run"); code != 6 {
		t.Fatal("unverified passed")
	}
	f.setRunState("running")
	if code := call("run", "wait", "--id", "run", "--wait-timeout", "5ms", "--poll-interval", "1ms"); code != 5 {
		t.Fatal(code, out.String())
	}
	if code := call("task", "run", "--id", "task", "--environment", "other"); code != 2 {
		t.Fatal("task config silently overridden")
	}
}
func TestConfigStrictAndExpiredSession(t *testing.T) {
	_, c, file := fixture(t)
	if err := os.WriteFile(file, []byte("base_url: http://localhost\nunknown: true\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := config.Load(file); err == nil {
		t.Fatal("unknown config accepted")
	}
	c.Config.Auth.ExpiresAt = time.Now().Add(-time.Second)
	if _, err := c.Call(context.Background(), "identity", "Session", map[string]any{}); clierror.From(err).Code != "auth" {
		t.Fatal(err)
	}
}

func (f *testPlatform) writeCount() int { f.mu.Lock(); defer f.mu.Unlock(); return f.writes }
func (f *testPlatform) asset(id string) map[string]any {
	f.mu.Lock()
	defer f.mu.Unlock()
	return jsonutil.Clone(f.assets[id])
}
func (f *testPlatform) setRunState(state string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.runState = state
}

func TestUncertainUpdateRestoresCreationDefaults(t *testing.T) {
	f, c, file := fixture(t)
	s := plan.Syncer{Client: c, ConfigFile: file}
	p := samplePlan()
	state, err := s.Sync(context.Background(), p)
	if err != nil {
		t.Fatal(err)
	}
	p.Entries[3].Name = "changed"
	f.mu.Lock()
	f.dropKind = "api"
	f.mu.Unlock()
	if _, err = s.Sync(context.Background(), p); clierror.From(err).Code != "uncertain" {
		t.Fatal(err)
	}
	if _, err = s.Bind(context.Background(), p, "api", state.Entries["api"].ID); err != nil {
		t.Fatal(err)
	}
	if _, err = s.Sync(context.Background(), p); err != nil {
		t.Fatal(err)
	}
	if f.writeCount() != 5 {
		t.Fatal("update was written again", f.writeCount())
	}
	if _, ok := p.Entries[3].Document["runner_id"]; ok {
		t.Fatal("sync mutated source plan")
	}
}
