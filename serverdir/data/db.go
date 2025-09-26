package data

import (
	"log"
	"os"
	"time"

	"github.com/Sahmaykf/GOstudy/serverdir/model"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

var DB *gorm.DB

func InitDB() (*gorm.DB, error) {
	dsn := "root:Hyfhgz87370376%40@tcp(127.0.0.1:3306)/GO-IM?parseTime=true&charset=utf8mb4&loc=Local"
	//emoji + 自动匹配时间
	gormLogger := logger.New(
		log.New(os.Stdout, "", log.LstdFlags),
		logger.Config{
			SlowThreshold:             500 * time.Millisecond,
			LogLevel:                  logger.Info,
			IgnoreRecordNotFoundError: true,
			Colorful:                  true,
		},
	)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		PrepareStmt: true, //预编译
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true, //单数表
		},
		Logger: gormLogger,
	})
	if err != nil {
		return nil, err
	}

	// 配置连接池
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxIdleConns(10)                  // 空闲连接上限
	sqlDB.SetMaxOpenConns(100)                 // 同时连接上限
	sqlDB.SetConnMaxLifetime(60 * time.Minute) // 连接寿命
	sqlDB.SetConnMaxIdleTime(10 * time.Minute) // 空闲寿命

	if err := db.AutoMigrate(&model.Account{}); err != nil {
		return nil, err
	}

	DB = db
	return db, nil
}
