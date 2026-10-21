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
	OrgID     string                  `json:"org_id"`
	ProjectID string                  `json:"project_id"`
	BatchID   string                  `json:"id"`
	Selected  []string                `json:"selected" validate:"required,min=1"`
	Updates   map[string]ImportUpdate `json:"updates"`
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
		for _, entry := range plan.Entries {
			if !selected[entry.TempID] {
				continue
			}
			ids[entry.TempID] = p.NewBase().ID
			if u, ok := in.Updates[entry.TempID]; ok {
				a, e := LoadResource(deps, tx, in.OrgID, in.ProjectID, u.ID, entry.Kind)
				if e != nil {
					return e
				}
				if a.Revision != u.Revision {
					return p.Conflict("更新目标已修改")
				}
				ids[entry.TempID] = u.ID
			}
		}
		if len(ids) != len(selected) {
			return p.Bad("选择项不存在")
		}
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
			folder := ""
			if entry.Folder != "" {
				folder = ids[entry.Folder]
				if folder == "" {
					missing = true
				}
			}
			if missing {
				return p.Bad("请选择引用资源及其父目录")
			}
			if e := validateDocument(deps, tx, &ResourceInput{OrgId: in.OrgID, ProjectId: in.ProjectID}, entry.Kind, doc); e != nil {
				return e
			}
			body, _ := json.Marshal(doc)
			if len(body) > 2<<20 {
				return p.Bad("单项超过2MiB")
			}
			record := model.Resource{Base: p.NewBase(), OrgID: in.OrgID, ProjectID: in.ProjectID, Kind: entry.Kind, FolderID: folder, Name: entry.Name, Revision: 1, Document: p.Seal(deps.Config.EncryptionKey, string(body))}
			record.ID = ids[entry.TempID]
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
