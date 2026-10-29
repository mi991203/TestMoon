package jsonutil

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	"github.com/mi991203/TestMoon/tools/apimoon/clierror"
)

func Decode(b []byte, v any) error {
	d := json.NewDecoder(bytes.NewReader(b))
	d.UseNumber()
	d.DisallowUnknownFields()
	if err := d.Decode(v); err != nil {
		return clierror.New("validation", "JSON 格式或字段类型不正确: "+err.Error())
	}
	var extra any
	if err := d.Decode(&extra); err != io.EOF {
		return clierror.New("validation", "输入必须是单份 JSON 文档")
	}
	return nil
}
func Object(v any) map[string]any { m, _ := v.(map[string]any); return m }
func String(v any) string {
	if v == nil {
		return ""
	}
	return fmt.Sprint(v)
}
func Clone(m map[string]any) map[string]any {
	b, _ := json.Marshal(m)
	var out map[string]any
	_ = Decode(b, &out)
	return out
}

func Merge(a, b map[string]any) map[string]any {
	out := Clone(a)
	for k, v := range b {
		out[k] = v
	}
	return out
}
