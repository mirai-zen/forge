package svc

import (
	"fmt"

	"github.com/mirai-zen/forge/platform/internal/config"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type ServiceContext struct {
	Config config.Config
	DB     *gorm.DB
}

func NewServiceContext(c config.Config) *ServiceContext {
	db, err := gorm.Open(mysql.Open(c.MySQL.DataSource), &gorm.Config{})
	if err != nil {
		panic(fmt.Sprintf("failed to connect database (dsn=%q): %v", c.MySQL.DataSource, err))
	}

	return &ServiceContext{
		Config: c,
		DB:     db,
	}
}
