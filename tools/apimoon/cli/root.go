package cli

import (
	"github.com/mi991203/TestMoon/libs/go/buildinfo"
	"github.com/mi991203/TestMoon/tools/apimoon/clierror"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func (a *application) rootCommand() *cobra.Command {
	o := &a.options
	root := commandGroup("apimoon", "AI Agent 调用 TestMoon 的 CLI")
	root.Long = "APIMoon：将接口、用例、场景和任务同步到 TestMoon，启动测试并读取报告。\n业务命令默认输出单份 JSON，进度写 stderr；帮助和 Shell 补全输出文本。"
	root.SilenceErrors = true
	root.SilenceUsage = true
	root.PersistentFlags().StringVar(&o.config, "config", "", "配置文件路径（默认 ~/.api_moon/config.yaml）")
	root.PersistentFlags().StringVar(&o.output, "output", "json", "业务结果格式：json 或 text（缩进 JSON）")
	root.PersistentFlags().StringVar(&o.org, "org", "", "临时选择组织 ID；切换组织会清除下级默认值")
	root.PersistentFlags().StringVar(&o.project, "project", "", "临时选择项目 ID；切换项目会清除环境和 Runner 默认值")
	root.Flags().BoolVar(&o.version, "version", false, "显示版本信息")
	root.PersistentPreRunE = func(cmd *cobra.Command, _ []string) error {
		if o.output != "json" && o.output != "text" {
			return clierror.New("usage", "output 必须为 json/text")
		}
		o.set = map[string]bool{}
		cmd.Flags().Visit(func(f *pflag.Flag) { o.set[f.Name] = true })
		return nil
	}
	root.RunE = func(cmd *cobra.Command, _ []string) error {
		if o.version {
			a.handled, a.data = true, buildinfo.Current()
			return nil
		}
		return cmd.Help()
	}
	root.AddCommand(a.command("version", "显示版本和构建信息", func(*cobra.Command, []string) (any, error) {
		return buildinfo.Current(), nil
	}))
	root.AddCommand(a.configCommand(), a.authCommand(), a.planCommand(), a.runCommand(), a.skillCommand())
	root.AddCommand(a.contextCommand(), a.orgCommand(), a.projectCommand(), a.runnerCommand())
	for _, kind := range []string{"api", "case", "scenario", "task", "environment", "folder"} {
		root.AddCommand(a.resourceCommand(kind))
	}
	return root
}
