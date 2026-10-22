package dao

import (
	"encoding/json"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/mi991203/TestMoon/backend/testing/mock/model"
	p "github.com/mi991203/TestMoon/libs/go/platform"
)

type Rules struct{ key string }

func NewRules(key string) *Rules { return &Rules{key: key} }

func (s *Rules) fields(value model.Rule) (map[string]any, error) {
	raw, err := p.Open(s.key, value.Document)
	if err != nil {
		return nil, err
	}
	var doc model.Response
	if err = json.Unmarshal([]byte(raw), &doc); err != nil {
		return nil, err
	}
	headers, err := json.Marshal(doc.Headers)
	if err != nil {
		return nil, err
	}
	conditions, err := json.Marshal(doc.Conditions)
	if err != nil {
		return nil, err
	}
	return map[string]any{"status": doc.Status, "delay_ms": doc.DelayMS, "headers_cipher": p.Seal(s.key, string(headers)), "body_cipher": p.Seal(s.key, doc.Body), "conditions_cipher": p.Seal(s.key, string(conditions)), "definition_version": 1}, nil
}

func (s *Rules) Response(value model.Rule) (model.Response, error) {
	result := model.Response{Status: value.Status, DelayMS: value.DelayMS}
	headers, err := p.Open(s.key, value.HeadersCipher)
	if err != nil {
		return result, err
	}
	if err = json.Unmarshal([]byte(headers), &result.Headers); err != nil {
		return result, err
	}
	conditions, err := p.Open(s.key, value.ConditionsCipher)
	if err != nil {
		return result, err
	}
	if err = json.Unmarshal([]byte(conditions), &result.Conditions); err != nil {
		return result, err
	}
	result.Body, err = p.Open(s.key, value.BodyCipher)
	return result, err
}

func (s *Rules) Create(db *gorm.DB, value *model.Rule) error {
	fields, err := s.fields(*value)
	if err != nil {
		return err
	}
	value.Status = fields["status"].(int)
	value.DelayMS = fields["delay_ms"].(int)
	value.HeadersCipher = fields["headers_cipher"].(string)
	value.BodyCipher = fields["body_cipher"].(string)
	value.ConditionsCipher = fields["conditions_cipher"].(string)
	value.DefinitionVersion = 1
	return db.Create(value).Error
}

func (s *Rules) Update(db *gorm.DB, value model.Rule) error {
	fields, err := s.fields(value)
	if err != nil {
		return err
	}
	fields["endpoint_id"] = value.EndpointID
	fields["name"] = value.Name
	fields["method"] = value.Method
	fields["path"] = value.Path
	fields["priority"] = value.Priority
	result := db.Model(&model.Rule{}).Where("org_id = ? AND project_id = ? AND id = ?", value.OrgID, value.ProjectID, value.ID).Updates(fields)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return p.NotFound()
	}
	return nil
}

func (s *Rules) Migrate(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		var rows []model.Rule
		return tx.Unscoped().Where("definition_version = 0").Clauses(clause.Locking{Strength: "UPDATE"}).FindInBatches(&rows, 100, func(_ *gorm.DB, _ int) error {
			for _, row := range rows {
				row.Document = row.LegacyDocument
				fields, err := s.fields(row)
				if err != nil {
					return fmt.Errorf("migrate mock definition: %w", err)
				}
				if err = tx.Unscoped().Model(&model.Rule{}).Where("id = ? AND definition_version = 0", row.ID).UpdateColumns(fields).Error; err != nil {
					return err
				}
			}
			return nil
		}).Error
	})
}
