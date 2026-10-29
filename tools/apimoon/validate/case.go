package validate

import (
	"fmt"
	"regexp"
	"slices"
	"strings"

	domain "github.com/mi991203/TestMoon/libs/go/testing/domain"
	"github.com/mi991203/TestMoon/tools/apimoon/clierror"
)

func validateCase(c domain.Case) error {
	if c.APIID == "" {
		if err := validateRequest(c.Request); err != nil {
			return err
		}
	}
	if err := validatePatch(c.Override); err != nil {
		return err
	}
	for _, ops := range [][]domain.Operation{c.Before, c.After} {
		if len(ops) > 20 {
			return clierror.New("validation", "前后置操作最多 20 项")
		}
		for _, op := range ops {
			if op.Type != "javascript" || op.Version != 1 || len(op.Script) > 65536 {
				return clierror.New("validation", "仅支持 version=1 的 javascript，脚本上限 64KiB")
			}
		}
	}
	for _, a := range c.Assertions {
		if !slices.Contains([]string{"status", "duration", "text", "header", "json"}, a.Source) || !slices.Contains([]string{"eq", "neq", "contains", "exists", "regex", "lt", "lte", "gt", "gte"}, a.Operator) {
			return clierror.New("validation", "断言 source/operator 不受支持")
		}
		if a.Source == "json" && a.Path != "" && !strings.HasPrefix(a.Path, "/") {
			return clierror.New("validation", "JSON 路径必须使用 JSON Pointer")
		}
		if a.Operator == "regex" {
			s := fmt.Sprint(a.Expected)
			if len(s) > 512 {
				return clierror.New("validation", "正则表达式超过 512 字节")
			}
			if _, err := regexp.Compile(s); err != nil {
				return clierror.New("validation", "无效正则表达式")
			}
		}
	}
	for _, x := range c.Extractions {
		if x.Variable == "" || !slices.Contains([]string{"status", "duration", "text", "header", "json"}, x.Source) {
			return clierror.New("validation", "提取配置缺少变量名或 source 不受支持")
		}
		if x.Source == "json" && x.Path != "" && !strings.HasPrefix(x.Path, "/") {
			return clierror.New("validation", "提取路径必须使用 JSON Pointer")
		}
	}
	return nil
}
