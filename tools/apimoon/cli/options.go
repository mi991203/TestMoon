package cli

import (
	"flag"
	"io"
	"slices"
	"strings"
	"time"

	"github.com/mi991203/TestMoon/tools/apimoon/clierror"
)

const help = `APIMoon — AI Agent 调用 TestMoon 的 Go CLI

apimoon config init --base-url URL
apimoon config show | config set KEY VALUE
apimoon auth login --username USER --password-stdin
apimoon auth status | auth logout | context show
apimoon org list | project list | runner list
apimoon {api|case|scenario|task|environment|folder} {list|get|create|update|delete}
apimoon plan {schema|validate|sync|status|bind} --file plan.json
apimoon task run --id TASK_ID [--wait] [--idempotency-key KEY]
apimoon run {list|get|wait|cancel} [--id RUN_ID]

通用: --config FILE --org ID --project ID --output json|text
资源: --file FILE|- --id ID --revision N --search TEXT --folder-id ID --api-id ID
创建默认: --environment ID --runner ID（也可在 config.yaml 中设置）
等待: --wait-timeout 30m --poll-interval 2s（超时不取消平台执行）
恢复绑定: plan bind --file plan.json --key KEY --id PLATFORM_ID

JSON 输入使用平台 snake_case 字段；更新必须提供 revision 和完整 document。
list|get|create|update|delete 输入可使用 --file，--file - 从 stdin 读取。
默认 stdout 为单份 JSON，进度写 stderr。plan schema 返回完整计划 JSON Schema。
退出码: 0成功，1请求/本地错误，2参数/校验，3鉴权/权限，4冲突/写入不明，5等待超时，6执行未通过。
`

type options struct {
	config, output, base, file, id, revision, org, project, environment, runner, search, folder, api, username, key, idempotency string
	wait, passwordStdin, help                                                                                                    bool
	waitTimeout, poll                                                                                                            time.Duration
	set                                                                                                                          map[string]bool
}

func parse(args []string) (options, []string, error) {
	o := options{set: map[string]bool{}}
	f := flag.NewFlagSet("apimoon", flag.ContinueOnError)
	f.SetOutput(io.Discard)
	for name, target := range map[string]*string{"config": &o.config, "output": &o.output, "base-url": &o.base, "file": &o.file, "id": &o.id, "revision": &o.revision, "org": &o.org, "project": &o.project, "environment": &o.environment, "runner": &o.runner, "search": &o.search, "folder-id": &o.folder, "api-id": &o.api, "username": &o.username, "key": &o.key, "idempotency-key": &o.idempotency} {
		f.StringVar(target, name, "", name)
	}
	f.BoolVar(&o.wait, "wait", false, "wait")
	f.BoolVar(&o.passwordStdin, "password-stdin", false, "password-stdin")
	f.BoolVar(&o.help, "help", false, "help")
	f.BoolVar(&o.help, "h", false, "help")
	f.DurationVar(&o.waitTimeout, "wait-timeout", 30*time.Minute, "wait-timeout")
	f.DurationVar(&o.poll, "poll-interval", 2*time.Second, "poll-interval")
	var flags, pos []string
	for i := 0; i < len(args); i++ {
		a := args[i]
		if !strings.HasPrefix(a, "-") || a == "-" {
			pos = append(pos, a)
			continue
		}
		name := strings.SplitN(strings.TrimLeft(a, "-"), "=", 2)[0]
		item := f.Lookup(name)
		if item == nil {
			return o, nil, clierror.New("usage", "未知选项: "+name)
		}
		flags = append(flags, a)
		if !strings.Contains(a, "=") {
			if bf, ok := item.Value.(interface{ IsBoolFlag() bool }); ok && bf.IsBoolFlag() {
				continue
			}
			i++
			if i >= len(args) {
				return o, nil, clierror.New("usage", "选项缺少值: "+name)
			}
			flags = append(flags, args[i])
		}
	}
	if err := f.Parse(flags); err != nil {
		return o, nil, clierror.New("usage", err.Error())
	}
	f.Visit(func(v *flag.Flag) { o.set[v.Name] = true })
	if o.output == "" {
		o.output = "json"
	}
	if !slices.Contains([]string{"json", "text"}, o.output) {
		return o, nil, clierror.New("usage", "output 必须为 json/text")
	}
	return o, pos, nil
}
