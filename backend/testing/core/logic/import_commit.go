package logic

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/mi991203/TestMoon/backend/testing/core/dao"
	"github.com/mi991203/TestMoon/backend/testing/core/internal/runtime"
	"github.com/mi991203/TestMoon/backend/testing/core/model"
	p "github.com/mi991203/TestMoon/libs/go/platform"
)

type ImportCommitInput struct {
	OrgID          string                  `json:"org_id"`
	ProjectID      string                  `json:"project_id"`
	BatchID        string                  `json:"id"`
	Selected       []string                `json:"selected" validate:"required,min=1"`
	Updates        map[string]ImportUpdate `json:"updates"`
	TargetFolderID string                  `json:"target_folder_id"`
}
type ImportUpdate struct {
	ID       string `json:"id"`
	Revision int64  `json:"revision"`
}

func CommitImport(ctx context.Context, deps *runtime.Dependencies, in ImportCommitInput) (int, error) {
	count := 0
	e := deps.DB.Transaction(func(tx *gorm.DB) error {
		var batch model.ImportBatch
		if e := dao.WithProjectScope(tx, in.OrgID, in.ProjectID).Clauses(clause.Locking{Strength: "UPDATE"}).First(&batch, "id = ? AND user_id = ?", in.BatchID, p.Actor(ctx).UserID).Error; e != nil {
			return e
		}
		if batch.Used || time.Now().After(batch.ExpiresAt) {
			return p.Conflict("导入预览已使用或过期")
		}
		raw, e := p.Open(deps.Config.EncryptionKey, batch.Plan)
		if e != nil {
			return e
		}
		var plan model.ImportPlan
		if json.Unmarshal([]byte(raw), &plan) != nil {
			return p.Bad("预览损坏")
		}
		selected := map[string]bool{}
		for _, id := range in.Selected {
			selected[id] = true
		}
		ids := map[string]string{}
		updateTargets := map[string]bool{}
		existingFolders := map[string]string{}
		existingDefaults := map[string]map[string]any{}
		if in.TargetFolderID != "" {
			target, err := LoadResource(deps, tx, in.OrgID, in.ProjectID, in.TargetFolderID, "folder")
			if err != nil {
				return err
			}
			doc, err := readResourceDocument(deps, target)
			if err != nil {
				return err
			}
			if doc["kind"] != "api" {
				return p.Bad("导入目标必须是API目录")
			}
		}
		for _, entry := range plan.Entries {
			if !selected[entry.TempID] {
				continue
			}
			ids[entry.TempID] = p.NewBase().ID
			if u, ok := in.Updates[entry.TempID]; ok {
				if updateTargets[u.ID] {
					return p.Bad("同一资源不能作为多个导入项的更新目标")
				}
				updateTargets[u.ID] = true
				a, e := LoadResource(deps, tx, in.OrgID, in.ProjectID, u.ID, entry.Kind)
				if e != nil {
					return e
				}
				if a.Revision != u.Revision {
					return p.Conflict("更新目标已修改")
				}
				ids[entry.TempID] = u.ID
				existingFolders[entry.TempID] = a.FolderID
				if entry.Kind == "api" || entry.Kind == "case" {
					doc, err := readResourceDocument(deps, a)
					if err != nil {
						return err
					}
					existingDefaults[entry.TempID] = doc
				}
				if entry.Kind == "folder" {
					doc, err := readResourceDocument(deps, a)
					if err != nil {
						return err
					}
					if doc["kind"] != entry.Document["kind"] {
						return p.Bad("更新目录类型不匹配")
					}
				}
			}
		}
		if len(ids) != len(selected) {
			return p.Bad("选择项不存在")
		}
		var records []model.Resource
		documents := map[string]map[string]any{}
		for _, entry := range plan.Entries {
			if !selected[entry.TempID] {
				continue
			}
			b, _ := json.Marshal(entry.Document)
			var d any
			_ = json.Unmarshal(b, &d)
			var missing bool
			var replace func(any) any
			replace = func(v any) any {
				switch x := v.(type) {
				case string:
					if strings.HasPrefix(x, "import:") {
						if id, ok := ids[x]; ok {
							return id
						}
						missing = true
					}
					return x
				case map[string]any:
					for k, v := range x {
						x[k] = replace(v)
					}
				case []any:
					for i, v := range x {
						x[i] = replace(v)
					}
				}
				return v
			}
			doc := replace(d).(map[string]any)
			if defaults, ok := existingDefaults[entry.TempID]; ok {
				// Imported formats do not carry the platform's execution defaults.
				doc["environment_id"] = defaults["environment_id"]
				doc["runner_id"] = defaults["runner_id"]
			}
			folder := ""
			if entry.Folder != "" {
				folder = ids[entry.Folder]
				if folder == "" {
					missing = true
				}
			}
			if existing, ok := existingFolders[entry.TempID]; ok {
				// An explicit update changes content, not the target's location.
				folder = existing
			} else if entry.Folder == "" && (entry.Kind == "api" || (entry.Kind == "folder" && doc["kind"] == "api")) {
				folder = in.TargetFolderID
			}
			if missing {
				return p.Bad("请选择引用资源及其父目录")
			}
			body, _ := json.Marshal(doc)
			if len(body) > 2<<20 {
				return p.Bad("单项超过2MiB")
			}
			record := model.Resource{Base: p.NewBase(), OrgID: in.OrgID, ProjectID: in.ProjectID, Kind: entry.Kind, FolderID: folder, Name: entry.Name, Revision: 1, Document: p.Seal(deps.Config.EncryptionKey, string(body))}
			record.ID = ids[entry.TempID]
			records = append(records, record)
			documents[record.ID] = doc
		}
		if err := validateImportFolders(deps, tx, in, records, documents); err != nil {
			return err
		}
		i := 0
		for _, entry := range plan.Entries {
			if !selected[entry.TempID] {
				continue
			}
			record := records[i]
			i++
			if e := validateDocument(deps, tx, &ResourceInput{OrgId: in.OrgID, ProjectId: in.ProjectID}, entry.Kind, documents[record.ID]); e != nil {
				return e
			}
			if u, ok := in.Updates[entry.TempID]; ok {
				if e := deps.Assets.Update(tx, &record, u.Revision); e != nil {
					return e
				}
			} else {
				if e := deps.Assets.Create(tx, &record); e != nil {
					return e
				}
			}
			count++
		}
		return tx.Model(&batch).Update("used", true).Error
	})
	if e != nil {
		return 0, CoreError(e)
	}
	return count, nil
}

// Validate the final parent graph before writing, including new folders and
// existing update targets, so imports obey the same type/depth rules as saves.
func validateImportFolders(deps *runtime.Dependencies, tx *gorm.DB, in ImportCommitInput, records []model.Resource, documents map[string]map[string]any) error {
	folders := map[string]model.Resource{}
	kinds := map[string]string{}
	for _, record := range records {
		if record.Kind == "folder" {
			folders[record.ID] = record
			kinds[record.ID], _ = documents[record.ID]["kind"].(string)
		}
	}
	for _, record := range records {
		want := record.Kind
		if want == "folder" {
			want = kinds[record.ID]
		}
		seen := map[string]bool{record.ID: true}
		parent := record.FolderID
		for depth := 0; parent != ""; depth++ {
			if seen[parent] || depth >= 32 {
				return p.Bad("目录不能循环或超过32层")
			}
			seen[parent] = true
			folder, ok := folders[parent]
			if !ok {
				var err error
				folder, err = LoadResource(deps, tx, in.OrgID, in.ProjectID, parent, "folder")
				if err != nil {
					return err
				}
				doc, err := readResourceDocument(deps, folder)
				if err != nil {
					return err
				}
				folders[parent] = folder
				kinds[parent], _ = doc["kind"].(string)
			}
			if kinds[parent] != want {
				return p.Bad("目录类型不匹配")
			}
			parent = folder.FolderID
		}
	}
	return nil
}
