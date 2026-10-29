package plan

import (
	"fmt"
	"slices"
	"sort"

	"github.com/mi991203/TestMoon/tools/apimoon/jsonutil"
)

type Reference struct{ Path, Kind, ID string }

// Visit only reference-bearing fields. Bodies, scripts, variables and JSON assertions
// are arbitrary user payload and must never participate in plan substitution.
func references(e Entry, visit func(Reference) (string, error)) error {
	var err error
	if e.FolderID != "" {
		_, err = visit(Reference{"folder_id", "folder", e.FolderID})
		if err != nil {
			return err
		}
	}
	var walk func(string, map[string]any, string) error
	walk = func(kind string, d map[string]any, prefix string) error {
		if d == nil {
			return nil
		}
		fields := map[string]string{}
		if slices.Contains([]string{"api", "case", "task"}, kind) {
			fields["environment_id"] = "environment"
		}
		if kind == "case" {
			fields["api_id"] = "api"
		}
		if kind == "step" {
			fields["case_id"] = "case"
		}
		if kind == "selection" && jsonutil.String(d["kind"]) != "folder" {
			fields["id"] = jsonutil.String(d["kind"])
		}
		keys := []string{}
		for key := range fields {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			if id := jsonutil.String(d[key]); id != "" {
				next, er := visit(Reference{prefix + key, fields[key], id})
				if er != nil {
					return er
				}
				d[key] = next
			}
		}
		if kind == "step" {
			if er := walk("case", jsonutil.Object(d["inline"]), prefix+"inline."); er != nil {
				return er
			}
		}
		for key, childKind := range map[string]string{"steps": "step", "children": kind, "items": "selection"} {
			if children, ok := d[key].([]any); ok {
				for i, child := range children {
					if er := walk(childKind, jsonutil.Object(child), fmt.Sprintf("%s%s[%d].", prefix, key, i)); er != nil {
						return er
					}
				}
			}
		}
		return nil
	}
	return walk(e.Kind, e.Document, "document.")
}
