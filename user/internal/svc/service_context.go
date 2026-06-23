// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package svc

import (
	"github.com/mirai-zen/forge/user/internal/config"
	"github.com/mirai-zen/forge/user/internal/model"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type ServiceContext struct {
	Config config.Config
	DB     *gorm.DB
}

func NewServiceContext(c config.Config) *ServiceContext {
	db := initDB(c)
	// Auto migrate
	if err := db.AutoMigrate(&model.User{}); err != nil {
		logx.Errorf("auto migrate user table failed: %v", err)
	}

	return &ServiceContext{
		Config: c,
		DB:     db,
	}
}

func initDB(c config.Config) *gorm.DB {
	db, err := gorm.Open(mysql.Open(c.MySQL.DataSource), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		logx.Errorf("connect mysql failed: %v", err)
		panic(err)
	}
	return db
}
