package model

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/codepository/user/config"
	"github.com/jinzhu/gorm"

	// mysql
	_ "github.com/go-sql-driver/mysql"
)

var db *gorm.DB

// Model 其它数据结构的公共部分
type Model struct {
	ID         int       `gorm:"primary_key" json:"id,omitempty"`
	CreateTime time.Time `gorm:"column:createTime" json:"createTime,omitempty"`
}

// 配置
var conf = *config.Config

// StartDB 启动数据库
func StartDB() {
	setupDB()
	setupWxdb()
}

// StopDB 关闭数据库
func StopDB() {
	CloseDB()
	closeWxDB()
	log.Println("关闭数据库")
}

// SetupDB 初始化一个db连接
func setupDB() {
	var err error
	db, err = gorm.Open(conf.DbType, fmt.Sprintf("%s:%s@(%s:%s)/%s?charset=utf8&parseTime=True&loc=Local", conf.DbUser, conf.DbPassword, conf.DbHost, conf.DbPort, conf.DbName))
	if err != nil {
		log.Fatalf("数据库连接失败 err: %v", err)
	}
	log.Println("启动数据库连接！！")
	// 启用Logger，显示详细日志
	mode, _ := strconv.ParseBool(conf.DbLogMode)

	db.LogMode(mode)

	db.SingularTable(true) //全局设置表名不可以为复数形式
	db.Callback().Create().Replace("gorm:update_time_stamp", updateTimeStampForCreateCallback)
	idle, err := strconv.Atoi(conf.DbMaxIdleConns)
	if err != nil {
		panic(err)
	}
	db.DB().SetMaxIdleConns(idle)
	db.DB().SetConnMaxLifetime(time.Hour * 2)
	open, err := strconv.Atoi(conf.DbMaxOpenConns)
	if err != nil {
		panic(err)
	}
	db.DB().SetMaxOpenConns(open)
	// gorm.DefaultTableNameHandler = func(db *gorm.DB, defaultTableName string) string {
	// 	return "fznews_" + defaultTableName
	// }
	db.Set("gorm:table_options", "ENGINE=InnoDB  DEFAULT CHARSET=utf8 AUTO_INCREMENT=1;").
		AutoMigrate(&FznewsLogin{}).AutoMigrate(&FznewsRecord{}).AutoMigrate(&FznewsLeadership{}).
		AutoMigrate(&UserOrg{}).AutoMigrate(&FznewsTask{}).AutoMigrate(&FznewsTaskUser{})
	db.Model(&FznewsLogin{}).AddUniqueIndex("user_id", "user_id")
	db.Model(&FznewsLeadership{}).Omit("department_name")
	db.Model(&FznewsLeadership{}).AddUniqueIndex("userid_departmentid_role", "userid", "department_id", "role")
	db.Model(&UserOrg{}).AddUniqueIndex("userid_orgid", "userid", "orgid")
	db.Model(&FznewsTask{}).AddUniqueIndex("name", "name")
	db.Model(&FznewsTaskUser{}).AddUniqueIndex("userid_task_start", "userid", "task", "start")
}

// CloseDB closes database connection (unnecessary)
func CloseDB() {
	defer db.Close()
}

// GetDB getdb
func GetDB() *gorm.DB {
	return db
}

// GetTx GetTx
func GetTx() *gorm.DB {
	return db.Begin()
}

// GetWxTx GetWxTx
func GetWxTx() *gorm.DB {
	return wxdb.Begin()
}
func updateTimeStampForCreateCallback(scope *gorm.Scope) {
	if !scope.HasError() {
		nowTime := time.Now()
		if createTimeField, ok := scope.FieldByName("CreateTime"); ok {
			if createTimeField.IsBlank {
				createTimeField.Set(nowTime)
			}
		}

		// if modifyTimeField, ok := scope.FieldByName("ModifiedOn"); ok {
		// 	if modifyTimeField.IsBlank {
		// 		modifyTimeField.Set(nowTime)
		// 	}
		// }
	}
}
