package execution

import (
	"context"
	"slices"
	"time"

	"github.com/mi991203/TestMoon/tools/apimoon/clierror"
	"github.com/mi991203/TestMoon/tools/apimoon/jsonutil"
	"github.com/mi991203/TestMoon/tools/apimoon/platform"
)

func Wait(ctx context.Context, c *platform.Client, id string, timeout, poll time.Duration) (map[string]any, error) {
	waitCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	var last map[string]any
	for {
		v, err := c.Call(waitCtx, "execution", "GetRun", jsonutil.Merge(c.Scope(), map[string]any{"id": id}))
		if err != nil {
			if waitCtx.Err() != nil {
				return waitError(ctx, c, id, last)
			}
			return last, err
		}
		last = v
		Decorate(c, v)
		state := jsonutil.String(jsonutil.Object(v["run"])["state"])
		if !slices.Contains([]string{"created", "queued", "running"}, state) {
			if state == "passed" {
				return v, nil
			}
			return v, clierror.New("execution", "执行结束，状态为 "+state)
		}
		timer := time.NewTimer(poll)
		select {
		case <-waitCtx.Done():
			timer.Stop()
			return waitError(ctx, c, id, last)
		case <-timer.C:
		}
	}
}
func waitError(ctx context.Context, c *platform.Client, id string, last map[string]any) (map[string]any, error) {
	if last == nil {
		last = map[string]any{"run": map[string]any{"id": id}, "url": c.Link("run", id)}
	}
	if ctx.Err() != nil {
		return last, clierror.New("interrupted", "等待已中断，平台执行未自动取消")
	}
	return last, clierror.New("wait_timeout", "等待超时，平台执行未自动取消，可继续 run wait")
}
