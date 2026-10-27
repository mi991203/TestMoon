package model

import "time"

// These fixture records are physically deleted so repeated test runs can clean up.
type Record struct {
	ID        string `gorm:"type:char(36);primaryKey"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

type School struct {
	Record
	Name    string `gorm:"type:varchar(200);not null"`
	Address string `gorm:"type:varchar(500);not null"`
}

type Class struct {
	Record
	SchoolID string `gorm:"type:char(36);not null;index"`
	Name     string `gorm:"type:varchar(200);not null"`
	School   School `gorm:"foreignKey:SchoolID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

type Student struct {
	Record
	ClassID   string `gorm:"type:char(36);not null;index"`
	Name      string `gorm:"type:varchar(200);not null"`
	StudentNo string `gorm:"type:varchar(100);not null"`
	Class     Class  `gorm:"foreignKey:ClassID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}
