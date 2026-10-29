package cli

import (
	"context"
	"io"

	"github.com/mi991203/TestMoon/tools/apimoon/clierror"
	"github.com/mi991203/TestMoon/tools/apimoon/jsonutil"
	"github.com/mi991203/TestMoon/tools/apimoon/platform"
	"github.com/mi991203/TestMoon/tools/apimoon/validate"
)

func resourceCommand(ctx context.Context, c *platform.Client, o options, group, action string, stdin io.Reader) (any, error) {
	input := map[string]any{}
	if o.file != "" {
		if err := readInput(o.file, stdin, &input); err != nil {
			return nil, err
		}
		if input == nil {
			return nil, clierror.New("validation", "输入必须是 JSON 对象")
		}
	}
	for key, value := range map[string]string{"id": o.id, "revision": o.revision, "search": o.search, "folder_id": o.folder, "api_id": o.api} {
		flagKey := map[string]string{"folder_id": "folder-id", "api_id": "api-id"}[key]
		if flagKey == "" {
			flagKey = key
		}
		if o.set[flagKey] {
			input[key] = value
		}
	}
	if action == "create" {
		c.Defaults(group, input)
	} else if action == "update" && (o.set["environment"] || o.set["runner"]) {
		return nil, clierror.New("usage", "更新请在完整 document 中明确填写环境和 Runner")
	}
	scoped, err := c.Scoped(input)
	if err != nil {
		return nil, err
	}
	if err = validate.Asset(group, action, scoped); err != nil {
		return nil, err
	}
	v, err := c.Resource(ctx, group, action, input)
	if err == nil {
		if r := jsonutil.Object(v["resource"]); r != nil {
			v["url"] = c.Link(group, jsonutil.String(r["id"]))
		}
	}
	return v, err
}
