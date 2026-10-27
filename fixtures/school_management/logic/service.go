package logic

import (
	"context"
	"errors"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/mi991203/TestMoon/fixtures/school_management/model"
)

type Service struct{ DB *gorm.DB }

func dbError(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, gorm.ErrRecordNotFound):
		return status.Error(codes.NotFound, "resource not found")
	case errors.Is(err, gorm.ErrForeignKeyViolated):
		return status.Error(codes.InvalidArgument, "referenced school or class does not exist")
	case errors.Is(err, context.Canceled):
		return status.Error(codes.Canceled, "request canceled")
	case errors.Is(err, context.DeadlineExceeded):
		return status.Error(codes.DeadlineExceeded, "request timed out")
	default:
		return status.Error(codes.Internal, "database operation failed")
	}
}

func validateText(field, value string, max int, required bool) error {
	if (required && strings.TrimSpace(value) == "") || utf8.RuneCountInString(value) > max {
		return status.Errorf(codes.InvalidArgument, "%s must %scontain at most %d characters", field, map[bool]string{true: "be non-empty and ", false: ""}[required], max)
	}
	return nil
}

func validateID(id string) error {
	if strings.TrimSpace(id) == "" {
		return status.Error(codes.InvalidArgument, "id is required")
	}
	return nil
}

func validateSchool(v model.School) error {
	if err := validateText("name", v.Name, 200, true); err != nil {
		return err
	}
	if err := validateText("address", v.Address, 500, false); err != nil {
		return err
	}
	return nil
}

func (s *Service) CreateSchool(ctx context.Context, v model.School) (model.School, error) {
	if err := validateSchool(v); err != nil {
		return v, err
	}
	v.Record = model.Record{ID: uuid.Must(uuid.NewV7()).String()}
	err := s.DB.WithContext(ctx).Omit(clause.Associations).Create(&v).Error
	return v, dbError(err)
}

func (s *Service) GetSchool(ctx context.Context, id string) (model.School, error) {
	var v model.School
	if err := validateID(id); err != nil {
		return v, err
	}
	err := s.DB.WithContext(ctx).Where("id = ?", id).First(&v).Error
	return v, dbError(err)
}

func (s *Service) UpdateSchool(ctx context.Context, v model.School) (model.School, error) {
	if err := validateID(v.ID); err != nil {
		return v, err
	}
	if err := validateSchool(v); err != nil {
		return v, err
	}
	// Lock the record so replacement and the returned representation are atomic.
	err := s.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var current model.School
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", v.ID).First(&current).Error; err != nil {
			return err
		}
		if err := tx.Model(&current).Updates(map[string]any{
			"name":    v.Name,
			"address": v.Address,
		}).Error; err != nil {
			return err
		}
		return tx.Where("id = ?", v.ID).First(&v).Error
	})
	return v, dbError(err)
}

func (s *Service) DeleteSchool(ctx context.Context, id string) error {
	if err := validateID(id); err != nil {
		return err
	}
	result := s.DB.WithContext(ctx).Where("id = ?", id).Delete(&model.School{})
	if result.Error != nil {
		return dbError(result.Error)
	}
	if result.RowsAffected == 0 {
		return dbError(gorm.ErrRecordNotFound)
	}
	return nil
}

func (s *Service) ListSchools(ctx context.Context) ([]model.School, error) {
	items := make([]model.School, 0)
	q := s.DB.WithContext(ctx).Model(&model.School{})
	err := q.Order("schools.created_at ASC").Order("schools.id ASC").Find(&items).Error
	return items, dbError(err)
}

func validateClass(v model.Class) error {
	if err := validateText("school_id", v.SchoolID, 36, true); err != nil {
		return err
	}
	if err := validateText("name", v.Name, 200, true); err != nil {
		return err
	}
	return nil
}

func (s *Service) CreateClass(ctx context.Context, v model.Class) (model.Class, error) {
	if err := validateClass(v); err != nil {
		return v, err
	}
	v.Record = model.Record{ID: uuid.Must(uuid.NewV7()).String()}
	err := s.DB.WithContext(ctx).Omit(clause.Associations).Create(&v).Error
	return v, dbError(err)
}

func (s *Service) GetClass(ctx context.Context, id string) (model.Class, error) {
	var v model.Class
	if err := validateID(id); err != nil {
		return v, err
	}
	err := s.DB.WithContext(ctx).Where("id = ?", id).First(&v).Error
	return v, dbError(err)
}

func (s *Service) UpdateClass(ctx context.Context, v model.Class) (model.Class, error) {
	if err := validateID(v.ID); err != nil {
		return v, err
	}
	if err := validateClass(v); err != nil {
		return v, err
	}
	// Lock the record so replacement and the returned representation are atomic.
	err := s.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var current model.Class
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", v.ID).First(&current).Error; err != nil {
			return err
		}
		if err := tx.Model(&current).Updates(map[string]any{
			"school_id": v.SchoolID,
			"name":      v.Name,
		}).Error; err != nil {
			return err
		}
		return tx.Where("id = ?", v.ID).First(&v).Error
	})
	return v, dbError(err)
}

func (s *Service) DeleteClass(ctx context.Context, id string) error {
	if err := validateID(id); err != nil {
		return err
	}
	result := s.DB.WithContext(ctx).Where("id = ?", id).Delete(&model.Class{})
	if result.Error != nil {
		return dbError(result.Error)
	}
	if result.RowsAffected == 0 {
		return dbError(gorm.ErrRecordNotFound)
	}
	return nil
}

func (s *Service) ListClasses(ctx context.Context, schoolID string) ([]model.Class, error) {
	items := make([]model.Class, 0)
	q := s.DB.WithContext(ctx).Model(&model.Class{})
	if schoolID != "" {
		q = q.Where("school_id = ?", schoolID)
	}
	err := q.Order("classes.created_at ASC").Order("classes.id ASC").Find(&items).Error
	return items, dbError(err)
}

func validateStudent(v model.Student) error {
	if err := validateText("class_id", v.ClassID, 36, true); err != nil {
		return err
	}
	if err := validateText("name", v.Name, 200, true); err != nil {
		return err
	}
	if err := validateText("student_no", v.StudentNo, 100, true); err != nil {
		return err
	}
	return nil
}

func (s *Service) CreateStudent(ctx context.Context, v model.Student) (model.Student, error) {
	if err := validateStudent(v); err != nil {
		return v, err
	}
	v.Record = model.Record{ID: uuid.Must(uuid.NewV7()).String()}
	err := s.DB.WithContext(ctx).Omit(clause.Associations).Create(&v).Error
	return v, dbError(err)
}

func (s *Service) GetStudent(ctx context.Context, id string) (model.Student, error) {
	var v model.Student
	if err := validateID(id); err != nil {
		return v, err
	}
	err := s.DB.WithContext(ctx).Where("id = ?", id).First(&v).Error
	return v, dbError(err)
}

func (s *Service) UpdateStudent(ctx context.Context, v model.Student) (model.Student, error) {
	if err := validateID(v.ID); err != nil {
		return v, err
	}
	if err := validateStudent(v); err != nil {
		return v, err
	}
	// Lock the record so replacement and the returned representation are atomic.
	err := s.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var current model.Student
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", v.ID).First(&current).Error; err != nil {
			return err
		}
		if err := tx.Model(&current).Updates(map[string]any{
			"class_id":   v.ClassID,
			"name":       v.Name,
			"student_no": v.StudentNo,
		}).Error; err != nil {
			return err
		}
		return tx.Where("id = ?", v.ID).First(&v).Error
	})
	return v, dbError(err)
}

func (s *Service) DeleteStudent(ctx context.Context, id string) error {
	if err := validateID(id); err != nil {
		return err
	}
	result := s.DB.WithContext(ctx).Where("id = ?", id).Delete(&model.Student{})
	if result.Error != nil {
		return dbError(result.Error)
	}
	if result.RowsAffected == 0 {
		return dbError(gorm.ErrRecordNotFound)
	}
	return nil
}

func (s *Service) ListStudents(ctx context.Context, schoolID, classID string) ([]model.Student, error) {
	items := make([]model.Student, 0)
	q := s.DB.WithContext(ctx).Model(&model.Student{})
	if classID != "" {
		q = q.Where("students.class_id = ?", classID)
	}
	if schoolID != "" {
		q = q.Joins("JOIN classes ON classes.id = students.class_id").Where("classes.school_id = ?", schoolID)
	}
	err := q.Order("students.created_at ASC").Order("students.id ASC").Find(&items).Error
	return items, dbError(err)
}
