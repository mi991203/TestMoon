package execution

import (
	"slices"

	"github.com/mi991203/TestMoon/tools/apimoon/jsonutil"
	"github.com/mi991203/TestMoon/tools/apimoon/platform"
)

func Decorate(c *platform.Client, v map[string]any) {
	v["url"] = c.Link("run", jsonutil.String(jsonutil.Object(v["run"])["id"]))
	counts := map[string]int{}
	var failures []any
	if items, ok := jsonutil.Object(v["result"])["items"].([]any); ok {
		for _, it := range items {
			item := jsonutil.Object(it)
			if steps, ok := item["steps"].([]any); ok {
				for _, st := range steps {
					step := jsonutil.Object(st)
					status := jsonutil.String(step["status"])
					counts[status]++
					if !slices.Contains([]string{"passed", "unverified"}, status) {
						failures = append(failures, map[string]any{"item_id": item["id"], "step_id": step["id"], "name": step["name"], "status": status, "error": step["error"], "assertions": step["assertions"]})
					}
				}
			}
		}
	}
	v["summary"] = counts
	if failures == nil {
		failures = []any{}
	}
	v["failures"] = failures
}
