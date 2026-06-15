package model

import (
	"encoding/json"
	"time"
)

type Service struct {
	ID         uint            `gorm:"primaryKey;autoIncrement"`
	ProjectID  uint            `gorm:"index;not null"`
	Name       string          `gorm:"size:64;not null"`
	Template   string          `gorm:"size:64;not null"`
	ParamsJSON json.RawMessage `gorm:"type:json;not null"`
	CreatedAt  time.Time       `gorm:"autoCreateTime"`
	UpdatedAt  time.Time       `gorm:"autoUpdateTime"`

	Project *Project     `gorm:"foreignKey:ProjectID"`
	Envs    []ServiceEnv `gorm:"foreignKey:ServiceID"`
}

func (Service) TableName() string {
	return "services"
}

type ServiceEnv struct {
	ID        uint      `gorm:"primaryKey;autoIncrement"`
	ServiceID uint      `gorm:"uniqueIndex:uk_env;not null"`
	Env       string    `gorm:"size:32;uniqueIndex:uk_env;not null"`
	Namespace string    `gorm:"size:64;not null"`
	CreatedAt time.Time `gorm:"autoCreateTime"`

	Service *Service `gorm:"foreignKey:ServiceID"`
}

func (ServiceEnv) TableName() string {
	return "service_envs"
}
