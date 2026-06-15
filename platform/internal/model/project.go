package model

import "time"

type Project struct {
	ID        uint      `gorm:"primaryKey;autoIncrement"`
	Name      string    `gorm:"size:64;uniqueIndex;not null"`
	GitOrg    string    `gorm:"size:128;not null"`
	GitRepo   string    `gorm:"size:128;not null"`
	Template  string    `gorm:"size:64;not null"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

func (Project) TableName() string {
	return "projects"
}
