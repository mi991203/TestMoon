package cli

import (
	"time"

	"github.com/spf13/cobra"
)

type options struct {
	agent, projectDir, directory                                                                                                 string
	global, force                                                                                                                bool
	config, output, base, file, id, revision, org, project, environment, runner, search, folder, api, username, key, idempotency string
	wait, passwordStdin, version                                                                                                 bool
	waitTimeout, poll                                                                                                            time.Duration
	set                                                                                                                          map[string]bool
}

func (o *options) inputFlag(cmd *cobra.Command) {
	cmd.Flags().StringVar(&o.file, "file", "", "输入 JSON 文件路径，- 表示从 stdin 读取")
}

func (o *options) idFlag(cmd *cobra.Command) {
	cmd.Flags().StringVar(&o.id, "id", "", "平台资源或执行 ID")
}

func (o *options) defaultFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&o.environment, "environment", "", "创建资源的默认环境 ID，显式空值清除配置默认值")
	cmd.Flags().StringVar(&o.runner, "runner", "", "创建资源的默认 Runner ID，显式空值清除配置默认值")
}

func (o *options) waitFlags(cmd *cobra.Command) {
	cmd.Flags().DurationVar(&o.waitTimeout, "wait-timeout", 30*time.Minute, "等待超时，超时不会取消平台执行")
	cmd.Flags().DurationVar(&o.poll, "poll-interval", 2*time.Second, "查询执行状态的轮询间隔")
}
