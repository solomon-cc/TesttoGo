package database

import (
	"fmt"
	"time"

	"testogo/internal/model/entity"
	"testogo/pkg/config"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

var DB *gorm.DB

func Init() error {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local&timeout=10s&writeTimeout=30s&readTimeout=30s",
		config.GetString("database.username"),
		config.GetString("database.password"),
		config.GetString("database.host"),
		config.GetInt("database.port"),
		config.GetString("database.dbname"),
		config.GetString("database.charset"),
	)

	db, err := gorm.Open(mysql.New(mysql.Config{
		DSN:               dsn,
		DefaultStringSize: 191, // MySQL 5.7 的 utf8mb4 索引长度限制
	}), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true, // 使用单数表名
		},
		DisableForeignKeyConstraintWhenMigrating: true, // 禁用外键约束
	})
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
	sqlDB.SetConnMaxLifetime(time.Hour) // 设置连接最大生命周期

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
