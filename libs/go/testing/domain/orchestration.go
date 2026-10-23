package domain

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

// Directories group nodes. Execution always visits leaves in depth-first order.
// Bound both depth and total nodes, including empty directories.
func flattenTree[T any](nodes []T, maxLeaves int, inspect func(T) (string, string, []T, error)) ([]T, error) {
	var leaves []T
	count := 0
	keys := map[string]bool{}
	var walk func([]T, int) error
	walk = func(nodes []T, depth int) error {
		if len(nodes) == 0 {
			return nil
		}
		if depth > 16 {
			return fmt.Errorf("编排最多16层")
		}
		for _, node := range nodes {
			count++
			if count > 1000 {
				return fmt.Errorf("编排节点最多1000个")
			}
			key, folderName, children, err := inspect(node)
			if err != nil {
				return err
			}
			if key != "" {
				if keys[key] {
					return fmt.Errorf("编排节点标识不能重复")
				}
				keys[key] = true
			}
			if folderName != "" {
				if strings.TrimSpace(folderName) == "" || utf8.RuneCountInString(folderName) > 128 {
					return fmt.Errorf("目录名称应为1..128个字符")
				}
				if err := walk(children, depth+1); err != nil {
					return err
				}
			} else {
				if len(children) != 0 {
					return fmt.Errorf("测试项必须是叶子节点")
				}
				leaves = append(leaves, node)
				if len(leaves) > maxLeaves {
					return fmt.Errorf("测试项数量超过限制（最多%d个）", maxLeaves)
				}
			}
		}
		return nil
	}
	if err := walk(nodes, 1); err != nil {
		return nil, err
	}
	return leaves, nil
}

func FlattenSelections(items []Selection, maxLeaves int) ([]Selection, error) {
	return flattenTree(items, maxLeaves, func(v Selection) (string, string, []Selection, error) {
		if v.Kind == "folder" {
			if v.Name == "" || v.ID != "" {
				return "", "", nil, fmt.Errorf("目录应有名称且不能引用测试项")
			}
			return v.NodeID, v.Name, v.Children, nil
		}
		if (v.Kind != "case" && v.Kind != "scenario") || v.ID == "" {
			return "", "", nil, fmt.Errorf("任务测试项应引用用例或场景")
		}
		return v.NodeID, "", v.Children, nil
	})
}

func FlattenSteps(steps []Step) ([]Step, error) {
	return flattenTree(steps, 100, func(v Step) (string, string, []Step, error) {
		if v.Kind == "folder" {
			if v.Name == "" || v.CaseID != "" || v.Inline != nil || len(v.Variables) != 0 || v.ContinueOnFailure {
				return "", "", nil, fmt.Errorf("场景目录只能包含名称和子节点")
			}
			return v.ID, v.Name, v.Children, nil
		}
		if v.Kind != "" && v.Kind != "step" {
			return "", "", nil, fmt.Errorf("无效的场景节点类型")
		}
		if (v.CaseID == "") == (v.Inline == nil) {
			return "", "", nil, fmt.Errorf("每个步骤应引用用例或包含独立请求")
		}
		return v.ID, "", v.Children, nil
	})
}
