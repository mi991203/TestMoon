package cli

import (
	"github.com/mi991203/TestMoon/tools/apimoon/skills"
	"github.com/spf13/cobra"
)

func (a *application) skillCommand() *cobra.Command {
	group := commandGroup("skill", "安装和检查 AI Agent Skill")
	group.Long = "Skill 内嵌在 APIMoon 中；安装和检查无需登录、网络或 TestMoon 配置。默认安装到当前项目。"
	install := a.skillActionCommand("install", "安装 Skill；更新时备份已有内容", skills.Installer.Install)
	install.Flags().BoolVar(&a.options.force, "force", false, "备份已有 Skill 后更新，保留本地修改的副本")
	install.Example = "  apimoon skill install --agent codex\n  apimoon skill install --agent trae --global\n  apimoon skill install --agent trae-cn --global"
	status := a.skillActionCommand("status", "检查 Skill 是否已安装及内容是否一致", skills.Installer.Status)
	group.AddCommand(install, status)
	return group
}

func (a *application) skillActionCommand(use, short string, run func(skills.Installer, skills.Options) (skills.Result, error)) *cobra.Command {
	cmd := a.command(use, short, func(*cobra.Command, []string) (any, error) {
		installer, err := skills.Local()
		if err != nil {
			return nil, err
		}
		o := a.options
		return run(installer, skills.Options{Agent: o.agent, Global: o.global, ProjectDir: o.projectDir, Directory: o.directory, Force: o.force})
	})
	o := &a.options
	cmd.Flags().StringVarP(&o.agent, "agent", "a", "", "目标 Agent：codex、trae 或 trae-cn")
	cmd.Flags().BoolVarP(&o.global, "global", "g", false, "安装到用户级 skills 目录")
	cmd.Flags().StringVar(&o.projectDir, "project-dir", "", "指定项目目录（默认当前目录）")
	cmd.Flags().StringVar(&o.directory, "dir", "", "自定义 skills 根目录")
	_ = cmd.MarkFlagRequired("agent")
	cmd.MarkFlagsMutuallyExclusive("global", "project-dir", "dir")
	return cmd
}
