package model

import (
	"fmt"
	"log"
	"strconv"
	"time"

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
	wxdb.DB().SetConnMaxLifetime(time.Hour * 2)
	idle, err := strconv.Atoi(conf.DbMaxIdleConns)
	if err != nil {
		panic(err)
	}
	wxdb.Set("gorm:table_options", "ENGINE=InnoDB  DEFAULT CHARSET=utf8 AUTO_INCREMENT=1;").
		AutoMigrate(&WeixinOauserTag{}).AutoMigrate(&FznewsFlowProcess{}).AutoMigrate(&WeixinFlowLog{}).AutoMigrate(&WeixinFlowApprovaldata{}).
		AutoMigrate(&WeixinLeaveDepartment{})
	wxdb.Model(&WeixinOauserTaguser{}).AddUniqueIndex("tagid_uid", "tagId", "uId")
	wxdb.Model(&WeixinOauserTag{}).AddUniqueIndex("tagName", "tagName")
	wxdb.Model(&FznewsFlowProcess{}).AddUniqueIndex("processInstanceId", "processInstanceId")
	wxdb.Model(&WeixinFlowApprovaldata{}).AddUniqueIndex("thirdNo", "thirdNo")
	wxdb.Model(&WeixinFlowLog{}).Omit("processname")
	wxdb.Model(&WeixinFlowLog{}).AddForeignKey("thirdNo", "fznews_flow_process(processInstanceId)", "CASCADE", "CASCADE")
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
