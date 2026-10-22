package model

import (
	p "github.com/mi991203/TestMoon/libs/go/platform"
)

type User struct {
	p.Base
	Username string `gorm:"size:64;uniqueIndex" json:"username"`
	Name     string `gorm:"size:128" json:"name"`
	Password string `gorm:"size:256" json:"-"`
	Admin    bool   `json:"admin"`
	Disabled bool   `json:"disabled"`
	Epoch    int64  `json:"-"`
}
type Organization struct {
	p.Base
	Name string `gorm:"size:128" json:"name"`
}
type Member struct {
	p.Base
	OrgID  string `gorm:"type:char(36);uniqueIndex:org_user" json:"org_id"`
	UserID string `gorm:"type:char(36);uniqueIndex:org_user" json:"user_id"`
	Role   string `gorm:"size:16" json:"role"`
}
