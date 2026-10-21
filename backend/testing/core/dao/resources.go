package dao

import (
	"encoding/json"
	"errors"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/mi991203/TestMoon/backend/testing/core/model"
	p "github.com/mi991203/TestMoon/libs/go/platform"
	m "github.com/mi991203/TestMoon/libs/go/testing/domain"
)

type Resources struct{ key string }

func NewResources(key string) *Resources { return &Resources{key: key} }

func (s *Resources) seal(v any) (string, error) {
	raw, e := json.Marshal(v)
	if e != nil {
		return "", e
	}
	return p.Seal(s.key, string(raw)), nil
}

func (s *Resources) open(value string, out any) error {
	if value == "" {
		return nil
	}
	raw, e := p.Open(s.key, value)
	if e != nil {
		return e
	}
	return json.Unmarshal([]byte(raw), out)
}

func owner(r model.Resource) model.DefinitionOwner {
	return model.DefinitionOwner{Base: r.Base, OrgID: r.OrgID, ProjectID: r.ProjectID}
}

func find(db *gorm.DB, r model.Resource, out any) error {
	return db.Where("id = ? AND org_id = ? AND project_id = ?", r.ID, r.OrgID, r.ProjectID).First(out).Error
}

func definition(kind string) any {
	switch kind {
	case "api":
		return &model.APIDefinition{}
	case "case":
		return &model.TestCaseDefinition{}
	case "scenario":
		return &model.ScenarioDefinition{}
	case "environment":
		return &model.EnvironmentDefinition{}
	case "folder":
		return &model.ResourceFolder{}
	}
	return nil
}

func (s *Resources) writeDefinition(db *gorm.DB, r model.Resource) error {
	var doc map[string]any
	if e := s.open(r.Document, &doc); e != nil {
		return e
	}
	var record any
	var err error
	seal := func(v any) string {
		if err != nil {
			return ""
		}
		var value string
		value, err = s.seal(v)
		return value
	}
	switch r.Kind {
	case "api":
		var v m.API
		raw, _ := json.Marshal(doc)
		if e := json.Unmarshal(raw, &v); e != nil {
			return e
		}
		record = &model.APIDefinition{DefinitionOwner: owner(r), Protocol: v.Protocol, Description: v.Description, RequestCipher: seal(v.Request), SchemaCipher: seal(v.Schema)}
	case "case":
		var v m.Case
		raw, _ := json.Marshal(doc)
		if e := json.Unmarshal(raw, &v); e != nil {
			return e
		}
		record = &model.TestCaseDefinition{DefinitionOwner: owner(r), Name: v.Name, APIID: v.APIID, RequestCipher: seal(v.Request), OverrideCipher: seal(v.Override), VariablesCipher: seal(v.Variables), CollectionVariablesCipher: seal(v.CollectionVariables), BeforeCipher: seal(v.Before), AfterCipher: seal(v.After), AssertionsCipher: seal(v.Assertions), ExtractionsCipher: seal(v.Extractions)}
	case "scenario":
		var v m.Scenario
		raw, _ := json.Marshal(doc)
		if e := json.Unmarshal(raw, &v); e != nil {
			return e
		}
		record = &model.ScenarioDefinition{DefinitionOwner: owner(r), Name: v.Name, StepsCipher: seal(v.Steps), CollectionVariablesCipher: seal(v.CollectionVariables)}
	case "environment":
		var v m.Environment
		raw, _ := json.Marshal(doc)
		if e := json.Unmarshal(raw, &v); e != nil {
			return e
		}
		record = &model.EnvironmentDefinition{DefinitionOwner: owner(r), VariablesCipher: seal(v.Variables)}
	case "folder":
		kind, _ := doc["kind"].(string)
		record = &model.ResourceFolder{DefinitionOwner: owner(r), AssetKind: kind}
	default:
		return fmt.Errorf("unknown resource kind %q", r.Kind)
	}
	if err != nil {
		return err
	}
	return db.Clauses(clause.OnConflict{UpdateAll: true}).Create(record).Error
}

func (s *Resources) Create(db *gorm.DB, r *model.Resource) error {
	if e := db.Create(r).Error; e != nil {
		return e
	}
	return s.writeDefinition(db, *r)
}

func (s *Resources) Update(db *gorm.DB, r *model.Resource, revision int64) error {
	var current model.Resource
	if e := find(db, *r, &current); e != nil {
		return e
	}
	if current.Kind != r.Kind {
		return p.Bad("资源类型不匹配")
	}
	result := db.Model(&model.Resource{}).Where("id = ? AND org_id = ? AND project_id = ? AND revision = ?", r.ID, r.OrgID, r.ProjectID, revision).Updates(map[string]any{"name": r.Name, "folder_id": r.FolderID, "revision": gorm.Expr("revision + 1")})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected != 1 {
		return p.Conflict("资源已被修改，请重新加载")
	}
	if err := find(db, *r, &current); err != nil {
		return err
	}
	r.Base = current.Base
	r.Revision = revision + 1
	return s.writeDefinition(db, *r)
}

func (s *Resources) Delete(db *gorm.DB, r model.Resource) error {
	payload := definition(r.Kind)
	if payload == nil {
		return p.Bad("无效资源类型")
	}
	if e := db.Where("id = ? AND org_id = ? AND project_id = ?", r.ID, r.OrgID, r.ProjectID).Delete(payload).Error; e != nil {
		return e
	}
	return db.Delete(&r).Error
}

func (s *Resources) Document(db *gorm.DB, r model.Resource) (map[string]any, error) {
	var value any
	var err error
	open := func(cipher string, out any) {
		if err == nil {
			err = s.open(cipher, out)
		}
	}
	switch r.Kind {
	case "api":
		var row model.APIDefinition
		if e := find(db, r, &row); e != nil {
			return nil, e
		}
		v := m.API{Protocol: row.Protocol, Description: row.Description}
		open(row.RequestCipher, &v.Request)
		open(row.SchemaCipher, &v.Schema)
		value = v
	case "case":
		var row model.TestCaseDefinition
		if e := find(db, r, &row); e != nil {
			return nil, e
		}
		v := m.Case{Name: row.Name, APIID: row.APIID}
		open(row.RequestCipher, &v.Request)
		open(row.OverrideCipher, &v.Override)
		open(row.VariablesCipher, &v.Variables)
		open(row.CollectionVariablesCipher, &v.CollectionVariables)
		open(row.BeforeCipher, &v.Before)
		open(row.AfterCipher, &v.After)
		open(row.AssertionsCipher, &v.Assertions)
		open(row.ExtractionsCipher, &v.Extractions)
		value = v
	case "scenario":
		var row model.ScenarioDefinition
		if e := find(db, r, &row); e != nil {
			return nil, e
		}
		v := m.Scenario{Name: row.Name}
		open(row.StepsCipher, &v.Steps)
		open(row.CollectionVariablesCipher, &v.CollectionVariables)
		value = v
	case "environment":
		var row model.EnvironmentDefinition
		if e := find(db, r, &row); e != nil {
			return nil, e
		}
		v := m.Environment{}
		open(row.VariablesCipher, &v.Variables)
		value = v
	case "folder":
		var row model.ResourceFolder
		if e := find(db, r, &row); e != nil {
			return nil, e
		}
		value = map[string]string{"kind": row.AssetKind}
	default:
		return nil, p.Bad("无效资源类型")
	}
	if err != nil {
		return nil, err
	}
	raw, e := json.Marshal(value)
	if e != nil {
		return nil, e
	}
	var out map[string]any
	e = json.Unmarshal(raw, &out)
	return out, e
}

// Hydrate runs on the caller's transaction so snapshots see matching metadata
// and definition revisions. The transient ciphertext is never written to SQL.
func (s *Resources) Hydrate(db *gorm.DB, r *model.Resource) error {
	doc, e := s.Document(db, *r)
	if e != nil {
		return e
	}
	r.Document, e = s.seal(doc)
	return e
}

func (s *Resources) Migrate(db *gorm.DB) error {
	const version = "typed_resource_definitions_v1"
	if e := db.AutoMigrate(&model.APIDefinition{}, &model.TestCaseDefinition{}, &model.ScenarioDefinition{}, &model.EnvironmentDefinition{}, &model.ResourceFolder{}, &model.SchemaMigration{}); e != nil {
		return e
	}
	err := db.Transaction(func(tx *gorm.DB) error {
		marker := model.SchemaMigration{Version: version}
		if e := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&marker).Error; e != nil {
			return e
		}
		var locked model.SchemaMigration
		if e := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&locked, "version = ?", version).Error; e != nil {
			return e
		}
		// Completed markers have a suffix; locking serializes concurrent migrations.
		var done int64
		if e := tx.Model(&model.SchemaMigration{}).Where("version = ?", version+"_done").Count(&done).Error; e != nil {
			return e
		}
		if done > 0 {
			return nil
		}
		var rows []model.Resource
		e := tx.Unscoped().Order("id").FindInBatches(&rows, 100, func(_ *gorm.DB, _ int) error {
			for _, row := range rows {
				target := definition(row.Kind)
				if target == nil {
					return fmt.Errorf("unknown legacy resource kind %q", row.Kind)
				}
				var count int64
				if e := tx.Unscoped().Model(target).Where("id = ?", row.ID).Count(&count).Error; e != nil {
					return e
				}
				if count > 0 {
					continue
				}
				if row.LegacyDocument == "" {
					return errors.New("legacy resource has no definition")
				}
				row.Document = row.LegacyDocument
				if e := s.writeDefinition(tx, row); e != nil {
					return fmt.Errorf("migrate %s definition: %w", row.Kind, e)
				}
			}
			return nil
		}).Error
		if e != nil {
			return e
		}
		return tx.Create(&model.SchemaMigration{Version: version + "_done"}).Error
	})
	if err != nil {
		return err
	}
	return s.migrateDefinitionNames(db)
}

// CaseFilter uses the indexed relationship, including legacy standalone cases.
func (s *Resources) CaseFilter(q *gorm.DB, org, project, apiID string) *gorm.DB {
	q = q.Joins("JOIN test_case_definitions AS case_definition ON case_definition.id = resources.id AND case_definition.org_id = resources.org_id AND case_definition.project_id = resources.project_id AND case_definition.deleted_at = 0")
	if apiID != "" {
		return q.Where("case_definition.api_id = ?", apiID)
	}
	return q.Where("case_definition.api_id = '' OR NOT EXISTS (SELECT 1 FROM resources AS parent_api WHERE parent_api.id = case_definition.api_id AND parent_api.org_id = ? AND parent_api.project_id = ? AND parent_api.kind = 'api' AND parent_api.deleted_at = 0)", org, project)
}

// migrateDefinitionNames preserves the nested names carried by imported cases
// and scenarios in installations upgraded before those columns were introduced.
func (s *Resources) migrateDefinitionNames(db *gorm.DB) error {
	const version = "typed_definition_names_v1"
	return db.Transaction(func(tx *gorm.DB) error {
		marker := model.SchemaMigration{Version: version}
		if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&marker).Error; err != nil {
			return err
		}
		var locked model.SchemaMigration
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&locked, "version = ?", version).Error; err != nil {
			return err
		}
		var completed int64
		if err := tx.Model(&model.SchemaMigration{}).Where("version = ?", version+"_done").Count(&completed).Error; err != nil {
			return err
		}
		if completed > 0 {
			return nil
		}
		var rows []model.Resource
		if err := tx.Unscoped().Where("kind IN ? AND document <> ''", []string{"case", "scenario"}).FindInBatches(&rows, 100, func(_ *gorm.DB, _ int) error {
			for _, row := range rows {
				var legacy struct {
					Name string `json:"name"`
				}
				if err := s.open(row.LegacyDocument, &legacy); err != nil {
					return err
				}
				if legacy.Name == "" {
					continue
				}
				if err := tx.Unscoped().Model(definition(row.Kind)).Where("id = ? AND org_id = ? AND project_id = ? AND name = ''", row.ID, row.OrgID, row.ProjectID).UpdateColumn("name", legacy.Name).Error; err != nil {
					return err
				}
			}
			return nil
		}).Error; err != nil {
			return err
		}
		return tx.Create(&model.SchemaMigration{Version: version + "_done"}).Error
	})
}
