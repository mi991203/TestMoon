// Package skills installs the bundled APIMoon Skill into supported agent directories.
package skills

import (
	"bytes"
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/mi991203/TestMoon/libs/go/buildinfo"
	"github.com/mi991203/TestMoon/tools/apimoon/clierror"
)

//go:embed apimoon
var bundled embed.FS

type Options struct {
	Agent      string
	Global     bool
	ProjectDir string
	Directory  string
	Force      bool
}

type Installer struct {
	Home       string
	WorkingDir string
}

type Result struct {
	Skill          string `json:"skill"`
	Agent          string `json:"agent"`
	Scope          string `json:"scope"`
	Path           string `json:"path"`
	Status         string `json:"status"`
	BundledVersion string `json:"bundled_version"`
	Backup         string `json:"backup,omitempty"`
}

func Local() (Installer, error) {
	home, _ := os.UserHomeDir()
	cwd, err := os.Getwd()
	return Installer{Home: home, WorkingDir: cwd}, err
}

func (i Installer) target(o Options) (Result, error) {
	agent := strings.ToLower(o.Agent)
	var project, global string
	switch agent {
	case "codex":
		project, global = ".agents", ".agents"
	case "trae":
		project, global = ".trae", ".trae"
	case "trae-cn":
		project, global = ".trae", ".trae-cn"
	default:
		return Result{}, clierror.New("usage", "--agent 必须为 codex、trae 或 trae-cn")
	}
	if (o.Global && o.ProjectDir != "") || (o.Directory != "" && (o.Global || o.ProjectDir != "")) {
		return Result{}, clierror.New("usage", "--global、--project-dir、--dir 不能同时使用")
	}
	scope, dir := "project", o.ProjectDir
	if dir == "" {
		dir = i.WorkingDir
	}
	root := filepath.Join(dir, project, "skills")
	if o.Global {
		if i.Home == "" {
			return Result{}, clierror.New("local", "无法确定用户目录")
		}
		scope, root = "global", filepath.Join(i.Home, global, "skills")
	}
	if o.Directory != "" {
		scope, root = "custom", o.Directory
	}
	if !filepath.IsAbs(root) {
		root = filepath.Join(i.WorkingDir, root)
	}
	path, err := filepath.Abs(filepath.Join(root, "apimoon"))
	if err != nil {
		return Result{}, err
	}
	return Result{Skill: "apimoon", Agent: agent, Scope: scope, Path: path, BundledVersion: buildinfo.Version}, nil
}

func contents() (map[string][]byte, error) {
	files := map[string][]byte{}
	err := fs.WalkDir(bundled, "apimoon", func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		data, err := bundled.ReadFile(path)
		if err == nil {
			files[strings.TrimPrefix(path, "apimoon/")] = data
		}
		return err
	})
	return files, err
}

func inspect(path string, files map[string][]byte) (string, error) {
	info, err := os.Lstat(path)
	if os.IsNotExist(err) {
		return "not_installed", nil
	}
	if err != nil {
		return "", err
	}
	if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return "different", nil
	}
	matched, count := true, 0
	err = filepath.WalkDir(path, func(name string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		count++
		rel, err := filepath.Rel(path, name)
		if err != nil {
			return err
		}
		expected, ok := files[filepath.ToSlash(rel)]
		if !ok || d.Type()&os.ModeSymlink != 0 || !d.Type().IsRegular() {
			matched = false
			return nil
		}
		actual, err := os.ReadFile(name)
		if err == nil && !bytes.Equal(actual, expected) {
			matched = false
		}
		return err
	})
	if err != nil {
		return "", err
	}
	if matched && count == len(files) {
		return "current", nil
	}
	return "different", nil
}

func (i Installer) Status(o Options) (Result, error) {
	r, err := i.target(o)
	if err != nil {
		return r, err
	}
	files, err := contents()
	if err == nil {
		r.Status, err = inspect(r.Path, files)
	}
	return r, err
}

func (i Installer) Install(o Options) (Result, error) {
	r, err := i.target(o)
	if err != nil {
		return r, err
	}
	root := filepath.Dir(r.Path)
	if err = os.MkdirAll(root, 0755); err != nil {
		return r, err
	}
	lock := filepath.Join(root, ".apimoon-install.lock")
	f, err := os.OpenFile(lock, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		if os.IsExist(err) {
			return r, clierror.New("conflict", "Skill 正在安装；异常退出后请确认没有安装进程再移除锁: "+lock)
		}
		return r, err
	}
	_, _ = fmt.Fprintf(f, "pid=%d\n", os.Getpid())
	_ = f.Close()
	defer os.Remove(lock)
	files, err := contents()
	if err != nil {
		return r, err
	}
	r.Status, err = inspect(r.Path, files)
	if err != nil {
		return r, err
	}
	if r.Status == "current" {
		r.Status = "unchanged"
		return r, nil
	}
	if r.Status == "different" {
		if !o.Force {
			return r, clierror.New("conflict", "已有 Skill 内容不同；比较后用 --force 备份并更新: "+r.Path)
		}
		info, err := os.Lstat(r.Path)
		if err != nil {
			return r, err
		}
		if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
			return r, clierror.New("conflict", "目标不是普通目录，请先处理文件或符号链接: "+r.Path)
		}
	}
	stage, err := os.MkdirTemp(root, ".apimoon-stage-")
	if err != nil {
		return r, err
	}
	defer os.RemoveAll(stage)
	for name, data := range files {
		target := filepath.Join(stage, filepath.FromSlash(name))
		if err = os.MkdirAll(filepath.Dir(target), 0755); err != nil {
			return r, err
		}
		if err = os.WriteFile(target, data, 0644); err != nil {
			return r, err
		}
	}
	if err = os.Chmod(stage, 0755); err != nil {
		return r, err
	}
	if r.Status == "different" {
		// Keep backups outside the skills root so agents do not discover duplicates.
		backupRoot, err := os.MkdirTemp(filepath.Dir(root), ".apimoon-skill-backup-")
		if err != nil {
			return r, err
		}
		backup := filepath.Join(backupRoot, "apimoon")
		if err = os.Rename(r.Path, backup); err != nil {
			_ = os.Remove(backupRoot)
			return r, err
		}
		r.Backup = backup
	}
	if err = os.Rename(stage, r.Path); err != nil {
		if r.Backup != "" {
			if restoreErr := os.Rename(r.Backup, r.Path); restoreErr != nil {
				return r, fmt.Errorf("安装失败 (%v)，原内容保存在 %s，恢复失败: %w", err, r.Backup, restoreErr)
			}
			_ = os.Remove(filepath.Dir(r.Backup))
			r.Backup = ""
		}
		return r, err
	}
	if r.Backup != "" {
		r.Status = "updated"
	} else {
		r.Status = "installed"
	}
	return r, nil
}
