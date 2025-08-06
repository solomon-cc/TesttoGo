package database

import (
	"fmt"

	"testogo/internal/model/entity"
	"testogo/pkg/config"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Init() error {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local",
		config.GetString("database.username"),
		config.GetString("database.password"),
		config.GetString("database.host"),
		config.GetInt("database.port"),
		config.GetString("database.dbname"),
		config.GetString("database.charset"),
	)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return err
	}

	// 设置连接池
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}

	sqlDB.SetMaxIdleConns(config.GetInt("database.maxIdleConns"))
	sqlDB.SetMaxOpenConns(config.GetInt("database.maxOpenConns"))

	// 自动迁移数据库表
	err = db.AutoMigrate(
		&entity.User{},
		&entity.Question{},
		&entity.Paper{},
		&entity.UserAnswer{},
	)
	if err != nil {
		return err
	}

	DB = db
	return nil
}
