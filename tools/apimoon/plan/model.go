package plan

import (
	"encoding/json"

	"github.com/mi991203/TestMoon/tools/apimoon/jsonutil"
)

type Plan struct {
	Version int         `json:"version"`
	PlanID  string      `json:"plan_id"`
	Source  *PlanSource `json:"source,omitempty"`
	Entries []Entry     `json:"entries"`
}
type PlanSource struct {
	Requirement string `json:"requirement,omitempty"`
	CodeRef     string `json:"code_ref,omitempty"`
}
type Entry struct {
	Key      string         `json:"key"`
	Kind     string         `json:"kind"`
	ID       string         `json:"id,omitempty"`
	Revision json.Number    `json:"revision,omitempty"`
	Name     string         `json:"name"`
	FolderID string         `json:"folder_id,omitempty"`
	Document map[string]any `json:"document"`
}

func (e Entry) input() map[string]any {
	in := map[string]any{"name": e.Name, "document": jsonutil.Clone(e.Document)}
	if e.FolderID != "" {
		in["folder_id"] = e.FolderID
	}
	if e.ID != "" {
		in["id"] = e.ID
		in["revision"] = string(e.Revision)
	}
	return in
}
