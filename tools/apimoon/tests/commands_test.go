package apimoon_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mi991203/TestMoon/tools/apimoon/cli"
	"github.com/mi991203/TestMoon/tools/apimoon/config"
	"github.com/mi991203/TestMoon/tools/apimoon/jsonutil"
)

func callJSON(t *testing.T, args []string, input string) (int, map[string]any, string) {
	t.Helper()
	var out, logs bytes.Buffer
	code := cli.Execute(context.Background(), args, strings.NewReader(input), &out, &logs)
	var reply map[string]any
	d := json.NewDecoder(&out)
	if err := d.Decode(&reply); err != nil {
		t.Fatalf("invalid JSON response: %v; %s", err, out.String())
	}
	var extra any
	if err := d.Decode(&extra); err != io.EOF {
		t.Fatalf("stdout contains more than one JSON response: %v", err)
	}
	return code, reply, logs.String()
}

func TestCommandHelpIsScopedAndOffline(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "missing.yaml")
	commands := [][]string{nil, {"version"}, {"config"}, {"config", "init"}, {"config", "show"}, {"config", "set"},
		{"auth"}, {"auth", "login"}, {"auth", "status"}, {"auth", "logout"}, {"context"}, {"context", "show"},
		{"org"}, {"org", "list"}, {"project"}, {"project", "list"}, {"runner"}, {"runner", "list"},
		{"plan"}, {"plan", "schema"}, {"plan", "validate"}, {"plan", "sync"}, {"plan", "status"}, {"plan", "bind"},
		{"run"}, {"run", "list"}, {"run", "get"}, {"run", "wait"}, {"run", "cancel"}, {"task", "run"},
		{"skill"}, {"skill", "install"}, {"skill", "status"}}
	for _, kind := range []string{"api", "case", "scenario", "task", "environment", "folder"} {
		commands = append(commands, []string{kind})
		for _, action := range []string{"list", "get", "create", "update", "delete"} {
			commands = append(commands, []string{kind, action})
		}
	}
	for _, command := range commands {
		t.Run(strings.Join(command, "/"), func(t *testing.T) {
			args := append([]string{"--config", missing}, command...)
			args = append(args, "--help")
			var out, logs bytes.Buffer
			if code := cli.Execute(context.Background(), args, nil, &out, &logs); code != 0 {
				t.Fatalf("help requires configuration: %d %s", code, out.String())
			}
			path := strings.TrimSpace("apimoon " + strings.Join(command, " "))
			if !strings.Contains(out.String(), "Usage:") || !strings.Contains(out.String(), path) || !strings.Contains(out.String(), "--config") || logs.Len() != 0 {
				t.Fatalf("invalid help: %s; stderr: %s", out.String(), logs.String())
			}
			if json.Valid(out.Bytes()) || strings.Contains(out.String(), `"ok":`) {
				t.Fatal("help must be plain text without a JSON trailer")
			}
		})
	}
	for _, tc := range []struct {
		args    []string
		want    string
		missing string
	}{
		{[]string{"plan", "sync", "-h"}, "--file", "--agent"},
		{[]string{"skill", "install", "-h"}, "--force", "--file"},
		{[]string{"skill", "status", "-h"}, "--agent", "--force"},
		{[]string{"task", "run", "-h"}, "--wait-timeout", "--environment"},
		{[]string{"task", "list", "-h"}, "--file", "--search"},
		{[]string{"api", "list", "-h"}, "--folder-id", "--api-id"},
		{[]string{"case", "list", "-h"}, "--api-id", "--revision"},
		{[]string{"help", "plan", "bind"}, "--key", "--agent"},
	} {
		var out, logs bytes.Buffer
		if code := cli.Execute(context.Background(), append(tc.args, "--config", missing), nil, &out, &logs); code != 0 ||
			!strings.Contains(out.String(), tc.want) || strings.Contains(out.String(), tc.missing) {
			t.Fatalf("incorrect command flags for %v: %d %s", tc.args, code, out.String())
		}
	}
}

func TestInvalidCommandsFailBeforeConfiguration(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "missing.yaml")
	for _, command := range [][]string{
		{"not-a-command"}, {"plan", "syncc"}, {"auth", "loginn"}, {"api", "unknown"}, {"skill", "remove"},
		{"version", "extra"}, {"context", "show", "extra"}, {"plan", "schema", "--file", "ignored.json"},
		{"config", "set", "base_url"}, {"config", "init"}, {"auth", "login", "--username", "tester"},
		{"plan", "sync"}, {"plan", "bind", "--file", "-"}, {"plan", "bind", "--file", "-", "--key=", "--id=x"},
		{"api", "create"}, {"case", "get"}, {"task", "update"}, {"api", "delete"},
		{"task", "run", "--id", "x", "--environment", "other"}, {"task", "run", "--id", "x", "extra"},
		{"run", "get"}, {"run", "get", "--id="}, {"run", "list", "--wait"},
		{"run", "wait", "--id", "x", "--poll-interval", "invalid"},
		{"run", "wait", "--id", "x", "--wait-timeout", "0s"},
		{"run", "wait", "--id", "x", "--poll-interval", "-1s"},
		{"skill", "status", "--agent", "codex", "--force"},
		{"skill", "install", "--agent", "codex", "--global", "--project-dir", "."},
		{"skill", "install", "--agent", "codex", "--dir", ".", "--project-dir", "."},
		{"skill", "install"}, {"version", "--output", "yaml"}, {"version", "--unexpected"},
	} {
		t.Run(strings.Join(command, " "), func(t *testing.T) {
			args := append([]string{"--config", missing}, command...)
			code, reply, logs := callJSON(t, args, "")
			if code != 2 || reply["ok"] != false || jsonutil.Object(reply["error"])["code"] != "usage" || logs != "" {
				t.Fatalf("expected single usage error before config loading: %d %#v; stderr: %s", code, reply, logs)
			}
		})
	}
}

func TestCompletionAndSchemaAreOffline(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "missing.yaml")
	for _, shell := range []string{"bash", "zsh", "fish", "powershell"} {
		var out, logs bytes.Buffer
		code := cli.Execute(context.Background(), []string{"completion", shell, "--config", missing}, nil, &out, &logs)
		if code != 0 || !strings.Contains(out.String(), "apimoon") || json.Valid(out.Bytes()) || logs.Len() != 0 {
			t.Fatalf("invalid completion for %s: %d %s", shell, code, out.String())
		}
	}
	code, reply, _ := callJSON(t, []string{"plan", "schema", "--config", missing}, "")
	if code != 0 || reply["ok"] != true || jsonutil.Object(reply["data"])["properties"] == nil {
		t.Fatal("schema requires configuration", code, reply)
	}
}

func TestScopeOverridesAndFlagPositions(t *testing.T) {
	file := filepath.Join(t.TempDir(), "config.yaml")
	c := config.Config{BaseURL: "https://example.test", HTTPTimeoutSeconds: 30,
		Context: config.Scope{OrgID: "org", ProjectID: "project", EnvironmentID: "env", RunnerID: "runner"}}
	if err := config.Save(file, c); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		args []string
		want config.Scope
	}{
		{[]string{"--config", file, "context", "show"}, c.Context},
		{[]string{"context", "--config", file, "show"}, c.Context},
		{[]string{"context", "show", "--config=" + file, "--org=other"}, config.Scope{OrgID: "other"}},
		{[]string{"--org", "org", "context", "show", "--config", file, "--project", "other"}, config.Scope{OrgID: "org", ProjectID: "other"}},
		{[]string{"context", "show", "--config", file, "--org="}, config.Scope{}},
		{[]string{"context", "show", "--config", file, "--project="}, config.Scope{OrgID: "org"}},
	} {
		code, reply, _ := callJSON(t, tc.args, "")
		want := jsonutil.Object(reply["data"])
		if code != 0 || want["org_id"] != tc.want.OrgID || want["project_id"] != tc.want.ProjectID ||
			jsonutil.String(want["environment_id"]) != tc.want.EnvironmentID || jsonutil.String(want["runner_id"]) != tc.want.RunnerID {
			t.Fatalf("scope override %v: %d %#v", tc.args, code, reply)
		}
	}
	saved, err := config.Load(file)
	if err != nil || saved.Context != c.Context {
		t.Fatal("temporary flags changed persisted config", saved.Context, err)
	}
	code, reply, _ := callJSON(t, []string{"config", "set", "--config", file, "context.runner_id", "--", "--runner-value"}, "")
	if code != 0 || reply["ok"] != true {
		t.Fatal("-- terminator was not respected", code, reply)
	}
	saved, _ = config.Load(file)
	if saved.Context.RunnerID != "--runner-value" {
		t.Fatal("flag-like positional value was lost")
	}
}

func TestResourceCLIInputDefaultsAndFilters(t *testing.T) {
	f, c, file := fixture(t)
	c.Config.Context.EnvironmentID = "saved-env"
	if err := config.Save(file, c.Config); err != nil {
		t.Fatal(err)
	}
	input := `{"name":"api","document":{"protocol":"http","request":{"method":"GET","url":"https://example.test/schools"}}}`
	for _, tc := range []struct {
		flags               []string
		environment, runner string
	}{
		{nil, "saved-env", "public"},
		{[]string{"--environment=", "--runner", ""}, "", ""},
		{[]string{"--environment", "other-env", "--runner=other-runner"}, "other-env", "other-runner"},
	} {
		args := append([]string{"api", "create", "--config", file, "--file", "-"}, tc.flags...)
		code, reply, _ := callJSON(t, args, input)
		id := jsonutil.String(jsonutil.Object(jsonutil.Object(reply["data"])["resource"])["id"])
		document := jsonutil.Object(f.asset(id)["document"])
		if code != 0 || jsonutil.String(document["environment_id"]) != tc.environment || jsonutil.String(document["runner_id"]) != tc.runner {
			t.Fatalf("creation defaults changed: %d %#v", code, reply)
		}
		// ID can be supplied in JSON; it must not become an always-required flag.
		code, reply, _ = callJSON(t, []string{"api", "get", "--config", file, "--file", "-"}, `{"id":"`+id+`"}`)
		if code != 0 || reply["ok"] != true {
			t.Fatal("JSON-only ID rejected", code, reply)
		}
		code, reply, _ = callJSON(t, []string{"api", "update", "--config", file, "--file", "-"},
			`{"id":"`+id+`","revision":1,"name":"updated","document":{"protocol":"http","request":{"method":"GET","url":"https://example.test/updated"}}}`)
		if code != 0 || reply["ok"] != true {
			t.Fatal("JSON-only revision rejected", code, reply)
		}
	}
	code, reply, _ := callJSON(t, []string{"case", "list", "--config", file, "--folder-id=", "--api-id="}, "")
	if code != 0 || reply["ok"] != true {
		t.Fatal("explicit empty optional filters rejected", code, reply)
	}
	saved, _ := config.Load(file)
	if saved.Context != c.Config.Context {
		t.Fatal("creation flags changed persisted defaults")
	}
}

func TestPlanCLIValidationSyncAndPartialFailure(t *testing.T) {
	f, _, file := fixture(t)
	p := samplePlan()
	b, _ := json.Marshal(p)
	for _, action := range []string{"validate", "sync", "status"} {
		code, reply, _ := callJSON(t, []string{"plan", action, "--config", file, "--file", "-"}, string(b))
		if code != 0 || reply["ok"] != true {
			t.Fatalf("%s: %d %#v", action, code, reply)
		}
	}
	if f.writeCount() != 4 {
		t.Fatal("CLI sync wrote unexpected resources", f.writeCount())
	}
	code, _, _ := callJSON(t, []string{"plan", "sync", "--config", file, "--file", "-"}, string(b))
	if code != 0 || f.writeCount() != 4 {
		t.Fatal("CLI sync lost replay state")
	}
	// A lost response must retain the journal in the error envelope.
	f.mu.Lock()
	f.dropKind = "api"
	f.mu.Unlock()
	p.PlanID = "another-plan"
	b, _ = json.Marshal(p)
	code, reply, _ := callJSON(t, []string{"plan", "sync", "--config", file, "--file", "-"}, string(b))
	if code != 4 || jsonutil.Object(reply["error"])["code"] != "uncertain" ||
		jsonutil.Object(jsonutil.Object(reply["data"])["state"])["entries"] == nil {
		t.Fatal("partial sync data was lost", code, reply)
	}
}

func TestLocalConfigInitAndTextResult(t *testing.T) {
	file := filepath.Join(t.TempDir(), "config.yaml")
	code, reply, _ := callJSON(t, []string{"config", "init", "--config", file, "--base-url", "https://example.test"}, "")
	if code != 0 || reply["ok"] != true {
		t.Fatal(code, reply)
	}
	code, reply, _ = callJSON(t, []string{"config", "init", "--config", file, "--base-url", "https://other.test"}, "")
	if code != 4 || jsonutil.Object(reply["error"])["code"] != "conflict" {
		t.Fatal("init overwrote existing configuration", code, reply)
	}
	info, err := os.Stat(file)
	if err != nil || info.Mode().Perm() != 0600 {
		t.Fatal("config permissions changed", info, err)
	}
	code, reply, _ = callJSON(t, []string{"config", "show", "--config", file, "--output", "text"}, "")
	if code != 0 || jsonutil.Object(reply["data"])["base_url"] != "https://example.test" {
		t.Fatal(code, reply)
	}
}
