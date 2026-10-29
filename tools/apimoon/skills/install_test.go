package skills

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mi991203/TestMoon/tools/apimoon/clierror"
)

func TestAgentPaths(t *testing.T) {
	base := t.TempDir()
	i := Installer{Home: filepath.Join(base, "user"), WorkingDir: filepath.Join(base, "project")}
	for _, tc := range []struct {
		agent, project, global string
	}{
		{"codex", ".agents", ".agents"},
		{"trae", ".trae", ".trae"},
		{"trae-cn", ".trae", ".trae-cn"},
	} {
		for _, global := range []bool{false, true} {
			r, err := i.target(Options{Agent: tc.agent, Global: global})
			if err != nil {
				t.Fatal(err)
			}
			root, dir := i.WorkingDir, tc.project
			if global {
				root, dir = i.Home, tc.global
			}
			if want := filepath.Join(root, dir, "skills", "apimoon"); r.Path != want {
				t.Fatalf("%s global=%v: %s != %s", tc.agent, global, r.Path, want)
			}
		}
	}
	for _, opts := range []Options{
		{Agent: "unknown"},
		{Agent: "codex", Global: true, ProjectDir: "project"},
		{Agent: "codex", Global: true, Directory: "skills"},
		{Agent: "codex", ProjectDir: "project", Directory: "skills"},
	} {
		if _, err := i.target(opts); err == nil || clierror.From(err).Code != "usage" {
			t.Fatalf("invalid options accepted: %+v %v", opts, err)
		}
	}
}

func TestInstallIdempotencyAndPreservingLocalChanges(t *testing.T) {
	i := Installer{Home: t.TempDir(), WorkingDir: t.TempDir()}
	opts := Options{Agent: "codex"}
	before, err := i.Status(opts)
	if err != nil || before.Status != "not_installed" {
		t.Fatal(before, err)
	}
	installed, err := i.Install(opts)
	if err != nil || installed.Status != "installed" {
		t.Fatal(installed, err)
	}
	for _, rel := range []string{"SKILL.md", "references/setup.md", "references/test-plans.md", "references/recovery.md", "agents/openai.yaml"} {
		if _, err := os.Stat(filepath.Join(installed.Path, rel)); err != nil {
			t.Fatalf("missing standalone resource %s: %v", rel, err)
		}
	}
	entry := filepath.Join(installed.Path, "SKILL.md")
	info, _ := os.Stat(entry)
	again, err := i.Install(opts)
	after, _ := os.Stat(entry)
	if err != nil || again.Status != "unchanged" || !info.ModTime().Equal(after.ModTime()) {
		t.Fatal(again, err)
	}
	if err = os.WriteFile(entry, []byte("local customization"), 0644); err != nil {
		t.Fatal(err)
	}
	if _, err = i.Install(opts); err == nil || clierror.From(err).Code != "conflict" {
		t.Fatal("overwrote local modifications", err)
	}
	content, _ := os.ReadFile(entry)
	if string(content) != "local customization" {
		t.Fatal("lost original skill")
	}
	opts.Force = true
	updated, err := i.Install(opts)
	if err != nil || updated.Status != "updated" || updated.Backup == "" {
		t.Fatal(updated, err)
	}
	content, _ = os.ReadFile(filepath.Join(updated.Backup, "SKILL.md"))
	if string(content) != "local customization" {
		t.Fatal("backup did not preserve local modifications")
	}
	rel, _ := filepath.Rel(filepath.Dir(updated.Path), updated.Backup)
	if !strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		t.Fatal("backup could be discovered as another skill", updated.Backup)
	}
	current, err := i.Status(opts)
	if err != nil || current.Status != "current" {
		t.Fatal(current, err)
	}
}

func TestInstallRejectsSymlinkAndConcurrentWriter(t *testing.T) {
	i := Installer{Home: t.TempDir(), WorkingDir: t.TempDir()}
	opts := Options{Agent: "trae-cn", Global: true, Force: true}
	target, _ := i.target(opts)
	if err := os.MkdirAll(filepath.Dir(target.Path), 0755); err != nil {
		t.Fatal(err)
	}
	outside := t.TempDir()
	if err := os.WriteFile(filepath.Join(outside, "SKILL.md"), []byte("keep me"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, target.Path); err != nil {
		t.Fatal(err)
	}
	if _, err := i.Install(opts); err == nil || clierror.From(err).Code != "conflict" {
		t.Fatal("followed skill symlink", err)
	}
	content, _ := os.ReadFile(filepath.Join(outside, "SKILL.md"))
	if string(content) != "keep me" {
		t.Fatal("modified symlink target")
	}
	if err := os.Remove(target.Path); err != nil {
		t.Fatal(err)
	}
	lock := filepath.Join(filepath.Dir(target.Path), ".apimoon-install.lock")
	if err := os.WriteFile(lock, []byte("pid=1"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := i.Install(opts); err == nil || clierror.From(err).Code != "conflict" {
		t.Fatal("ignored concurrent installation", err)
	}
}
