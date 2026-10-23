package domain

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestTaskTreeOrderAndRoundTrip(t *testing.T) {
	items := []Selection{
		{Kind: "folder", NodeID: "group", Name: "登录流程", Children: []Selection{
			{Kind: "case", ID: "login", NodeID: "first"},
			{Kind: "folder", NodeID: "nested", Name: "业务", Children: []Selection{{Kind: "scenario", ID: "checkout", NodeID: "second"}}},
			{Kind: "folder", Name: "空目录", NodeID: "empty"},
		}},
		{Kind: "case", ID: "login", NodeID: "third"},
	}
	raw, err := json.Marshal(Task{Items: items})
	if err != nil {
		t.Fatal(err)
	}
	var stored Task
	if err = json.Unmarshal(raw, &stored); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(stored.Items, items) {
		t.Fatal("tree changed during JSON persistence")
	}
	flat, err := FlattenSelections(stored.Items, 3)
	if err != nil {
		t.Fatal(err)
	}
	var got []string
	for _, v := range flat {
		got = append(got, v.NodeID)
	}
	if !reflect.DeepEqual(got, []string{"first", "second", "third"}) {
		t.Fatalf("execution order: %v", got)
	}
	if _, err = FlattenSelections(items, 2); err == nil {
		t.Fatal("nested leaves bypassed item limit")
	}
}

func TestScenarioTreePreservesStepConfiguration(t *testing.T) {
	body := "overridden"
	want := Step{Kind: "step", ID: "request", Inline: &Case{Name: "independent"}, Override: RequestPatch{Body: &body}, Variables: map[string]string{"token": "123"}, ContinueOnFailure: true}
	steps := []Step{{Kind: "folder", ID: "directory", Name: "请求目录", Children: []Step{want}}, {ID: "reference", CaseID: "case"}}
	flat, err := FlattenSteps(steps)
	if err != nil {
		t.Fatal(err)
	}
	if len(flat) != 2 || !reflect.DeepEqual(flat[0], want) || flat[1].CaseID != "case" {
		t.Fatalf("configuration lost: %+v", flat)
	}
	many := make([]Step, 101)
	for i := range many {
		many[i] = Step{CaseID: "case"}
	}
	if _, err := FlattenSteps([]Step{{Kind: "folder", Name: "目录", Children: many}}); err == nil {
		t.Fatal("nested steps bypassed 100-step limit")
	}
}

func TestInvalidTaskTrees(t *testing.T) {
	leaf := Selection{Kind: "case", ID: "case"}
	deep := []Selection{leaf}
	for i := 0; i < 16; i++ {
		deep = []Selection{{Kind: "folder", Name: "目录", Children: deep}}
	}
	oversized := make([]Selection, 1001)
	for i := range oversized {
		oversized[i] = Selection{Kind: "folder", Name: "空目录"}
	}
	tests := map[string][]Selection{
		"leaf with children":         {{Kind: "case", ID: "case", Children: []Selection{leaf}}},
		"folder with reference":      {{Kind: "folder", Name: "目录", ID: "case"}},
		"unnamed folder":             {{Kind: "folder"}},
		"blank folder":               {{Kind: "folder", Name: "  "}},
		"unsupported type":           {{Kind: "api", ID: "api"}},
		"missing reference":          {{Kind: "scenario"}},
		"duplicate node identity":    {{Kind: "folder", NodeID: "same", Name: "目录", Children: []Selection{{Kind: "case", ID: "case", NodeID: "same"}}}},
		"too deep":                   deep,
		"too many empty directories": oversized,
	}
	for name, tree := range tests {
		t.Run(name, func(t *testing.T) {
			if _, err := FlattenSelections(tree, 100); err == nil {
				t.Fatal("accepted invalid tree")
			}
		})
	}
}

func TestInvalidScenarioTrees(t *testing.T) {
	tests := map[string][]Step{
		"folder with request":   {{Kind: "folder", Name: "目录", Inline: &Case{}}},
		"folder with variables": {{Kind: "folder", Name: "目录", Variables: map[string]string{"x": "y"}}},
		"leaf with children":    {{CaseID: "case", Children: []Step{{CaseID: "child"}}}},
		"both request types":    {{CaseID: "case", Inline: &Case{}}},
		"missing request":       {{Kind: "step"}},
		"unknown kind":          {{Kind: "unknown", CaseID: "case"}},
		"duplicate step id":     {{ID: "same", CaseID: "case"}, {ID: "same", CaseID: "case"}},
	}
	for name, tree := range tests {
		t.Run(name, func(t *testing.T) {
			if _, err := FlattenSteps(tree); err == nil {
				t.Fatal("accepted invalid tree")
			}
		})
	}
}
