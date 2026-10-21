package logic

import (
	"context"
	"strings"

	"gorm.io/gorm"

	"github.com/mi991203/TestMoon/backend/testing/core/internal/runtime"
	"github.com/mi991203/TestMoon/backend/testing/core/model"
	p "github.com/mi991203/TestMoon/libs/go/platform"
)

func ListResources(ctx context.Context, deps *runtime.Dependencies, org, project, kind, search string, folderID, apiID *string, id string) ([]map[string]any, error) {
	_, role, err := RequireProject(ctx, deps.DB, org, project, "read")
	if err != nil {
		return nil, err
	}
	var items []map[string]any
	err = deps.DB.Transaction(func(tx *gorm.DB) error {
		var e error
		items, e = listResourceRows(deps, tx, org, project, kind, search, folderID, apiID, id, role)
		return e
	})
	return items, err
}

func listResourceRows(deps *runtime.Dependencies, tx *gorm.DB, org, project, kind, search string, folderID, apiID *string, id, role string) ([]map[string]any, error) {
	r := &ResourceInput{OrgId: org, ProjectId: project, Id: id}
	q := tx.Model(&model.Resource{}).Where("resources.org_id = ? AND resources.project_id = ? AND resources.kind = ?", org, project, kind)
	if id != "" {
		q = q.Where("resources.id = ?", id)
	}
	if search != "" {
		q = q.Where("resources.name LIKE ?", "%"+strings.ReplaceAll(strings.ReplaceAll(search, "%", "\\%"), "_", "\\_")+"%")
	}
	if folderID != nil {
		q = q.Where("resources.folder_id = ?", *folderID)
	}
	if apiID != nil {
		return casesForAPI(deps, tx, q, r, *apiID, role)
	}
	var rows []model.Resource
	if err := q.Order("name,id").Limit(1000).Find(&rows).Error; err != nil {
		return nil, p.DBError(err)
	}
	items := []map[string]any{}
	for _, row := range rows {
		if err := deps.Assets.Hydrate(tx, &row); err != nil {
			return nil, err
		}
		item, err := resourceView(deps, row, role)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}
