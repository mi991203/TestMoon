package importer

import (
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"sort"
	"strings"

	"github.com/pb33f/libopenapi"
	yaml "go.yaml.in/yaml/v3"

	"github.com/mi991203/TestMoon/backend/testing/core/internal/validation"
	"github.com/mi991203/TestMoon/backend/testing/core/model"
	p "github.com/mi991203/TestMoon/libs/go/platform"
	m "github.com/mi991203/TestMoon/libs/go/testing/domain"
)

func tempID() string { return "import:" + p.NewBase().ID }

func text(v any) string {
	if s, ok := v.(string); ok {
		return s
	}
	if v == nil {
		return ""
	}
	b, _ := json.Marshal(v)
	return string(b)
}

func object(v any) map[string]any {
	x, _ := v.(map[string]any)
	if x == nil {
		return map[string]any{}
	}
	return x
}

func array(v any) []any { x, _ := v.([]any); return x }

func addEntry(plan *ImportPlan, kind, name, folder string, v any) string {
	id := tempID()
	if len(name) > 128 {
		name = string([]rune(name)[:min(len([]rune(name)), 100)])
	}
	plan.Entries = append(plan.Entries, ImportEntry{TempID: id, Kind: kind, Name: name, Folder: folder, Document: jsonObject(v)})
	return id
}

func pmScripts(events any, source string) ([]m.Operation, []m.Operation) {
	var before, after []m.Operation
	for _, x := range array(events) {
		event := object(x)
		script := object(event["script"])
		lines := []string{}
		switch raw := script["exec"].(type) {
		case []any:
			for _, line := range raw {
				lines = append(lines, text(line))
			}
		case string:
			lines = []string{raw}
		}
		op := m.Operation{Type: "javascript", Version: 1, Script: strings.Join(lines, "\n"), Source: source}
		if event["listen"] == "prerequest" {
			before = append(before, op)
		} else if event["listen"] == "test" {
			after = append(after, op)
		}
	}
	return before, after
}

func Postman(content []byte) (ImportPlan, error) {
	plan := ImportPlan{}
	var root map[string]any
	if json.Unmarshal(content, &root) != nil {
		return plan, p.Bad("Postman文件不是有效JSON")
	}
	info := object(root["info"])
	if !strings.Contains(text(info["schema"]), "/v2.1.0/") {
		return plan, p.Bad("仅支持Postman Collection v2.1")
	}
	vars := map[string]string{}
	for _, x := range array(root["variable"]) {
		v := object(x)
		vars[text(v["key"])] = text(v["value"])
	}
	pre, post := pmScripts(root["event"], "collection")
	steps := []m.Step{}
	unsupported := regexp.MustCompile(`\b(require\s*\(|import\s|pm\.(execution|vault|visualizer|iterationData)|postman\.setNextRequest)`)
	var walk func([]any, string, string, []m.Operation, []m.Operation, any)
	walk = func(items []any, apiFolder, caseFolder string, parentPre, parentPost []m.Operation, parentAuth any) {
		for _, raw := range items {
			item := object(raw)
			name := text(item["name"])
			if name == "" {
				name = "导入请求"
			}
			bp, ap := pmScripts(item["event"], name)
			before := append(append([]m.Operation{}, parentPre...), bp...)
			after := append(append([]m.Operation{}, parentPost...), ap...)
			auth := parentAuth
			if item["auth"] != nil {
				auth = item["auth"]
			}
			if children, ok := item["item"].([]any); ok {
				af := addEntry(&plan, "folder", name, apiFolder, map[string]any{"kind": "api"})
				cf := addEntry(&plan, "folder", name, caseFolder, map[string]any{"kind": "case"})
				walk(children, af, cf, before, after, auth)
				continue
			}
			request := object(item["request"])
			if len(request) == 0 {
				plan.Diagnostics = append(plan.Diagnostics, name+": 无有效请求，已跳过")
				continue
			}
			if request["auth"] != nil {
				auth = request["auth"]
			}
			req := m.HTTPRequest{Method: strings.ToUpper(text(request["method"])), Headers: map[string]string{}, Query: map[string]string{}, BodyType: "text"}
			if u, ok := request["url"].(string); ok {
				req.URL = u
			} else {
				req.URL = text(object(request["url"])["raw"])
			}
			if req.URL == "" {
				plan.Diagnostics = append(plan.Diagnostics, name+": URL缺失，已跳过")
				continue
			}
			for _, x := range array(request["header"]) {
				h := object(x)
				if h["disabled"] != true {
					req.Headers[text(h["key"])] = text(h["value"])
				}
			}
			body := object(request["body"])
			switch body["mode"] {
			case "raw":
				req.Body = text(body["raw"])
				if object(object(body["options"])["raw"])["language"] == "json" {
					req.BodyType = "json"
				}
			case "urlencoded":
				req.BodyType = "form"
				values := url.Values{}
				for _, x := range array(body["urlencoded"]) {
					v := object(x)
					if v["disabled"] != true {
						values.Add(text(v["key"]), text(v["value"]))
					}
				}
				req.Body = values.Encode()
			case nil, "":
			default:
				plan.Diagnostics = append(plan.Diagnostics, name+": 请求体类型首版不支持，已跳过")
				continue
			}
			a := object(auth)
			switch a["type"] {
			case nil, "", "noauth":
			case "basic", "bearer", "apikey":
				values := map[string]string{}
				for _, x := range array(a[text(a["type"])]) {
					v := object(x)
					values[text(v["key"])] = text(v["value"])
				}
				if a["type"] == "basic" {
					req.Auth = m.Auth{Type: "basic", Username: values["username"], Password: values["password"]}
				} else if a["type"] == "bearer" {
					req.Auth = m.Auth{Type: "bearer", Token: values["token"]}
				} else {
					req.Auth = m.Auth{Type: "api_key", Key: values["key"], Value: values["value"], In: values["in"]}
				}
			default:
				plan.Diagnostics = append(plan.Diagnostics, name+": 认证类型不支持，已跳过")
				continue
			}
			if e := validation.Request(req); e != nil {
				plan.Diagnostics = append(plan.Diagnostics, name+": "+e.Error())
				continue
			}
			for _, op := range append(append([]m.Operation{}, before...), after...) {
				if unsupported.MatchString(op.Script) {
					plan.Diagnostics = append(plan.Diagnostics, name+": 脚本含不支持接口，运行时将明确失败，请修改")
				}
			}
			apiID := addEntry(&plan, "api", name, apiFolder, m.API{Protocol: "http", Request: req, Description: text(request["description"])})
			caseID := addEntry(&plan, "case", name, caseFolder, m.Case{APIID: apiID, Name: name, Before: before, After: after, CollectionVariables: vars})
			steps = append(steps, m.Step{ID: p.NewBase().ID, Name: name, CaseID: caseID})
		}
	}
	walk(array(root["item"]), "", "", pre, post, root["auth"])
	if len(steps) > 100 {
		return plan, p.Bad("导入集合超过100请求，请拆分集合")
	}
	if len(steps) > 0 {
		addEntry(&plan, "scenario", text(info["name"]), "", m.Scenario{Name: text(info["name"]), Steps: steps, CollectionVariables: vars})
	}
	return plan, nil
}

func OpenAPI(content []byte) (ImportPlan, error) {
	plan := ImportPlan{}
	if _, e := libopenapi.NewDocument(content); e != nil {
		return plan, p.Bad("OpenAPI解析失败: " + e.Error())
	}
	var root map[string]any
	if e := yaml.Unmarshal(content, &root); e != nil {
		return plan, p.Bad("无效JSON/YAML")
	}
	version := text(root["openapi"])
	swagger := text(root["swagger"])
	if swagger != "2.0" && !strings.HasPrefix(version, "3.0.") && !strings.HasPrefix(version, "3.1.") {
		return plan, p.Bad("仅支持OpenAPI 2.0/3.0/3.1")
	}
	var scan func(any) bool
	scan = func(v any) bool {
		switch x := v.(type) {
		case map[string]any:
			if ref, ok := x["$ref"].(string); ok && !strings.HasPrefix(ref, "#/") {
				return true
			}
			for _, v := range x {
				if scan(v) {
					return true
				}
			}
		case []any:
			for _, v := range x {
				if scan(v) {
					return true
				}
			}
		}
		return false
	}
	if scan(root) {
		return plan, p.Bad("首版仅支持文档内部 $ref，请先将外部引用打包为单文件")
	}
	base := "{{base_url}}"
	if servers := array(root["servers"]); len(servers) > 0 {
		base = strings.TrimRight(text(object(servers[0])["url"]), "/")
	} else if host := text(root["host"]); host != "" {
		scheme := "https"
		if ss := array(root["schemes"]); len(ss) > 0 {
			scheme = text(ss[0])
		}
		base = scheme + "://" + host + text(root["basePath"])
	}
	paths := object(root["paths"])
	keys := []string{}
	for k := range paths {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, path := range keys {
		item := resolveRef(root, object(paths[path]), 0)
		for _, method := range []string{"get", "post", "put", "patch", "delete", "head", "options"} {
			op, ok := item[method]
			if !ok {
				continue
			}
			operation := resolveRef(root, object(op), 0)
			if operation["security"] != nil || root["security"] != nil {
				plan.Diagnostics = append(plan.Diagnostics, path+": 安全方案保留在契约中，需手动配置 Basic/Bearer/API Key 认证")
			}
			if len(array(operation["servers"])) > 0 || len(array(item["servers"])) > 0 {
				plan.Diagnostics = append(plan.Diagnostics, path+": 首版使用文档首个全局 server，操作级 server 需手动调整")
			}

			name := text(operation["summary"])
			if name == "" {
				name = strings.ToUpper(method) + " " + path
			}
			req := m.HTTPRequest{Method: strings.ToUpper(method), URL: base + regexp.MustCompile(`\{([^{}]+)\}`).ReplaceAllString(path, "{{$1}}"), Headers: map[string]string{}, Query: map[string]string{}}
			parameters := append(append([]any{}, array(item["parameters"])...), array(operation["parameters"])...)
			for _, x := range parameters {
				v := resolveRef(root, object(x), 0)
				n := text(v["name"])
				sample := text(v["example"])
				if sample == "" {
					sample = text(v["default"])
				}
				switch v["in"] {
				case "query":
					req.Query[n] = sample
				case "header":
					req.Headers[n] = sample
				case "formData":
					plan.Diagnostics = append(plan.Diagnostics, name+": OpenAPI 2 formData 需手动补充，首版不转换上传文件")
				case "body":
					req.BodyType = "json"
					req.Body = "{}"
				}
			}
			rb := resolveRef(root, object(operation["requestBody"]), 0)
			if c := object(rb["content"]); len(c) > 0 {
				if data, ok := c["application/json"]; ok {
					req.BodyType = "json"
					req.Body = "{}"
					if ex := object(data)["example"]; ex != nil {
						req.Body = text(ex)
					}
				} else if _, ok := c["application/x-www-form-urlencoded"]; ok {
					req.BodyType = "form"
				} else if _, ok := c["text/plain"]; ok {
					req.BodyType = "text"
				} else {
					plan.Diagnostics = append(plan.Diagnostics, fmt.Sprintf("%s: 请求体格式首版不支持，已跳过", name))
					continue
				}
			}
			addEntry(&plan, "api", name, "", m.API{Protocol: "http", Request: req, Description: text(operation["description"]), Schema: operation})
		}
	}
	return plan, nil
}

func resolveRef(root, current map[string]any, depth int) map[string]any {
	if depth > 20 {
		return current
	}
	ref, _ := current["$ref"].(string)
	if !strings.HasPrefix(ref, "#/") {
		return current
	}
	var value any = root
	for _, part := range strings.Split(strings.TrimPrefix(ref, "#/"), "/") {
		key := strings.ReplaceAll(strings.ReplaceAll(part, "~1", "/"), "~0", "~")
		value = object(value)[key]
	}
	if value == nil {
		return current
	}
	return resolveRef(root, object(value), depth+1)
}

type ImportEntry = model.ImportEntry
type ImportPlan = model.ImportPlan

func jsonObject(v any) map[string]any {
	raw, _ := json.Marshal(v)
	var d map[string]any
	_ = json.Unmarshal(raw, &d)
	return d
}
