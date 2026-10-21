package logic

import (
	"gorm.io/gorm"

	"github.com/mi991203/TestMoon/backend/testing/core/internal/runtime"
	"github.com/mi991203/TestMoon/backend/testing/core/model"
	p "github.com/mi991203/TestMoon/libs/go/platform"
)

func casesForAPI(deps *runtime.Dependencies, tx, q *gorm.DB, r *ResourceInput, apiID, role string) ([]map[string]any, error) {
	if apiID != "" {
		if _, err := LoadResource(deps, tx, r.OrgId, r.ProjectId, apiID, "api"); err != nil {
			return nil, err
		}
	}
	var rows []model.Resource
	if err := deps.Assets.CaseFilter(q, r.OrgId, r.ProjectId, apiID).Select("resources.*").Order("resources.name, resources.id").Limit(1000).Find(&rows).Error; err != nil {
		return nil, p.DBError(err)
	}
	items := make([]map[string]any, 0, len(rows))
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
