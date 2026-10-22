package platform

import (
	"encoding/json"
	"net/url"
	"strings"
)

func Sanitize(v any, secrets []string) any {
	b, _ := json.Marshal(v)
	var data any
	_ = json.Unmarshal(b, &data)
	var walk func(any) any
	walk = func(v any) any {
		switch x := v.(type) {
		case map[string]any:
			secret, _ := x["secret"].(bool)
			kind, _ := x["type"].(string)
			path, _ := x["path"].(string)
			for k, v := range x {
				_, flag := v.(bool)
				if (Sensitive(k) && !(k == "secret" && flag)) || ((secret || kind == "api_key") && k == "value") || (k == "expected" && Sensitive(path)) {
					// Response headers retain their []string shape after redaction.
					if values, ok := v.([]any); ok {
						for i := range values {
							values[i] = "***"
						}
						x[k] = values
					} else {
						x[k] = "***"
					}
				} else {
					x[k] = walk(v)
				}
			}
			return x
		case []any:
			for i, v := range x {
				x[i] = walk(v)
			}
			return x
		case string:
			if u, err := url.Parse(x); err == nil && (u.Scheme == "http" || u.Scheme == "https") && u.Host != "" {
				q := u.Query()
				dirty := false
				for k := range q {
					if Sensitive(k) {
						q.Set(k, "***")
						dirty = true
					}
				}
				if u.User != nil {
					u.User = url.User("***")
					dirty = true
				}
				if dirty {
					u.RawQuery = q.Encode()
					x = u.String()
				}
			}
			for _, s := range secrets {
				if s != "" {
					x = strings.ReplaceAll(x, s, "***")
				}
			}
			if len(x) > 1 && (x[0] == '{' || x[0] == '[') {
				var nested any
				if json.Unmarshal([]byte(x), &nested) == nil {
					n, _ := json.Marshal(walk(nested))
					return string(n)
				}
			}
			return x
		default:
			return v
		}
	}
	return walk(data)
}
