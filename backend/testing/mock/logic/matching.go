package logic

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/mi991203/TestMoon/backend/testing/mock/internal/runtime"
	"github.com/mi991203/TestMoon/backend/testing/mock/model"
)

type MatchInput struct {
	Method, Path string
	Headers      http.Header
	Query        url.Values
	Body         any
}

func jsonPointer(v any, path string) (any, bool) {
	if path == "" {
		return v, true
	}
	if !strings.HasPrefix(path, "/") {
		return nil, false
	}
	for _, part := range strings.Split(path[1:], "/") {
		key := strings.ReplaceAll(strings.ReplaceAll(part, "~1", "/"), "~0", "~")
		if obj, ok := v.(map[string]any); ok {
			var exists bool
			v, exists = obj[key]
			if !exists {
				return nil, false
			}
		} else if array, ok := v.([]any); ok {
			index, err := strconv.Atoi(key)
			if err != nil || index < 0 || index >= len(array) {
				return nil, false
			}
			v = array[index]
		} else {
			return nil, false
		}
	}
	return v, true
}

// MatchResponse resolves priority and conditions without writing an HTTP response.
func MatchResponse(deps *runtime.Dependencies, endpoint model.Endpoint, input MatchInput) (*model.Response, error) {
	var rules []model.Rule
	if e := deps.DB.Where("org_id = ? AND project_id = ? AND endpoint_id = ? AND method = ? AND path = ?", endpoint.OrgID, endpoint.ProjectID, endpoint.ID, input.Method, input.Path).Find(&rules).Error; e != nil {
		return nil, e
	}
	sort.Slice(rules, func(i, j int) bool {
		if rules[i].Priority == rules[j].Priority {
			return rules[i].ID < rules[j].ID
		}
		return rules[i].Priority > rules[j].Priority
	})
	for _, rule := range rules {
		response, e := deps.Rules.Response(rule)
		if e != nil {
			return nil, e
		}
		matched := true
		for _, c := range response.Conditions {
			var got any
			exists := false
			switch c.Source {
			case "header":
				got = input.Headers.Get(c.Path)
				exists = input.Headers.Values(c.Path) != nil
			case "query":
				got = input.Query.Get(c.Path)
				_, exists = input.Query[c.Path]
			case "body":
				got, exists = jsonPointer(input.Body, c.Path)
			}
			ok := false
			switch c.Operator {
			case "exists":
				ok = exists
			case "eq":
				a, _ := json.Marshal(got)
				b, _ := json.Marshal(c.Expected)
				ok = exists && string(a) == string(b)
			case "contains":
				ok = exists && strings.Contains(fmt.Sprint(got), fmt.Sprint(c.Expected))
			case "regex":
				re, e := regexp.Compile(fmt.Sprint(c.Expected))
				ok = e == nil && exists && re.MatchString(fmt.Sprint(got))
			}
			if !ok {
				matched = false
				break
			}
		}
		if !matched {
			continue
		}
		return &response, nil
	}
	return nil, nil
}
