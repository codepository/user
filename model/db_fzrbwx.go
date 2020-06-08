package model

import (
	"fmt"
	"log"
	"strconv"

	"github.com/jinzhu/gorm"

	// mysql
	_ "github.com/go-sql-driver/mysql"
)

var wxdb *gorm.DB

// SetupWxdb 启动微信数据库
func setupWxdb() {
	var err error
	wxdb, err = gorm.Open(conf.DbType, fmt.Sprintf("%s:%s@(%s:%s)/%s?charset=utf8&parseTime=True&loc=Local",
		conf.WxDbUser, conf.WxDbPassword, conf.WxDbHost, conf.WxDbPort, conf.WxDbName))
	if err != nil {
		log.Fatalf("微信数据库连接失败 err: %v", err)
	}
	log.Println("启动微信数据库连接！！")
	// 启用Logger，显示详细日志
	mode, _ := strconv.ParseBool(conf.DbLogMode)
	wxdb.LogMode(mode)
	wxdb.SingularTable(true)
	idle, err := strconv.Atoi(conf.DbMaxIdleConns)
	if err != nil {
		panic(err)
	}
	wxdb.DB().SetMaxIdleConns(idle)
	open, err := strconv.Atoi(conf.DbMaxOpenConns)
	if err != nil {
		panic(err)
	}
	wxdb.DB().SetMaxOpenConns(open)

}
func closeWxDB() {
	defer wxdb.Close()
}
