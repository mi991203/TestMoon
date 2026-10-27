package dao

import (
	"context"
	"fmt"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/mi991203/TestMoon/fixtures/school_management/model"
)

func Open(ctx context.Context, dsn string) (*gorm.DB, error) {
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		TranslateError: true,
		Logger:         logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, fmt.Errorf("database connection failed")
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("database pool unavailable")
	}
	sqlDB.SetMaxOpenConns(10)
	sqlDB.SetMaxIdleConns(3)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)
	if err = sqlDB.PingContext(ctx); err == nil {
		err = db.WithContext(ctx).AutoMigrate(&model.School{}, &model.Class{}, &model.Student{})
	}
	if err != nil {
		_ = sqlDB.Close()
		return nil, fmt.Errorf("database initialization failed")
	}
	return db, nil
}
