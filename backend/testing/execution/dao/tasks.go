package dao

import (
	"encoding/json"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/mi991203/TestMoon/backend/testing/execution/model"
	p "github.com/mi991203/TestMoon/libs/go/platform"
	m "github.com/mi991203/TestMoon/libs/go/testing/domain"
)

type Tasks struct{ key string }

func NewTasks(key string) *Tasks { return &Tasks{key: key} }

func (s *Tasks) fields(value model.Task) (map[string]any, error) {
	raw, err := p.Open(s.key, value.Document)
	if err != nil {
		return nil, err
	}
	var doc m.Task
	if err = json.Unmarshal([]byte(raw), &doc); err != nil {
		return nil, err
	}
	items, err := json.Marshal(doc.Items)
	if err != nil {
		return nil, err
	}
	return map[string]any{"items_json": string(items), "environment_id": doc.EnvironmentID, "runner_id": doc.RunnerID, "timeout_seconds": doc.TimeoutSeconds, "config_version": 1}, nil
}

func (s *Tasks) Definition(value model.Task) (m.Task, error) {
	result := m.Task{EnvironmentID: value.EnvironmentID, RunnerID: value.RunnerID, TimeoutSeconds: value.TimeoutSeconds}
	err := json.Unmarshal([]byte(value.ItemsJSON), &result.Items)
	return result, err
}

func (s *Tasks) Create(db *gorm.DB, value *model.Task) error {
	fields, err := s.fields(*value)
	if err != nil {
		return err
	}
	value.ItemsJSON = fields["items_json"].(string)
	value.EnvironmentID = fields["environment_id"].(string)
	value.RunnerID = fields["runner_id"].(string)
	value.TimeoutSeconds = fields["timeout_seconds"].(int)
	value.ConfigVersion = 1
	return db.Create(value).Error
}

func (s *Tasks) Update(db *gorm.DB, value model.Task, revision int64) error {
	fields, err := s.fields(value)
	if err != nil {
		return err
	}
	fields["name"] = value.Name
	fields["revision"] = gorm.Expr("revision + 1")
	result := db.Model(&model.Task{}).Where("org_id = ? AND project_id = ? AND id = ? AND revision = ?", value.OrgID, value.ProjectID, value.ID, revision).Updates(fields)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected != 1 {
		return p.Conflict("任务已修改，请重新加载")
	}
	return nil
}

func (s *Tasks) Migrate(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		var rows []model.Task
		return tx.Unscoped().Where("config_version = 0").Clauses(clause.Locking{Strength: "UPDATE"}).FindInBatches(&rows, 100, func(_ *gorm.DB, _ int) error {
			for _, row := range rows {
				row.Document = row.LegacyDocument
				fields, err := s.fields(row)
				if err != nil {
					return fmt.Errorf("migrate task configuration: %w", err)
				}
				if err = tx.Unscoped().Model(&model.Task{}).Where("id = ? AND config_version = 0", row.ID).UpdateColumns(fields).Error; err != nil {
					return err
				}
			}
			return nil
		}).Error
	})
}
